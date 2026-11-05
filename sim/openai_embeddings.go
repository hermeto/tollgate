package sim

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/hermeto/tollgate/internal/tokens"
)

type oaiEmbeddingsRequest struct {
	Model          string          `json:"model"`
	Input          json.RawMessage `json:"input"`
	EncodingFormat string          `json:"encoding_format"`
	Dimensions     *int            `json:"dimensions"`
}

type oaiEmbedding struct {
	Object    string `json:"object"`
	Embedding any    `json:"embedding"`
	Index     int    `json:"index"`
}

func (s *Server) handleOpenAIEmbeddings(w http.ResponseWriter, r *http.Request) {
	body, ok := s.readBody(w, r)
	if !ok {
		return
	}
	p, ok := s.applyFaults(w, r, requestHash(r.URL.Path, body))
	if !ok {
		return
	}
	var req oaiEmbeddingsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "We could not parse the JSON body of your request: "+err.Error())
		return
	}
	if !s.models[req.Model] {
		s.writeError(w, http.StatusNotFound, errModelNotFound, fmt.Sprintf("The model `%s` does not exist or you do not have access to it.", req.Model))
		return
	}
	inputs, err := embeddingInputs(req.Input)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, err.Error())
		return
	}
	dims := s.cfg.Embeddings.Dimensions
	if req.Dimensions != nil {
		dims = *req.Dimensions
	}
	if dims < 1 || dims > maxEmbeddingDimensions {
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, fmt.Sprintf("dimensions must be between 1 and %d", maxEmbeddingDimensions))
		return
	}
	if req.EncodingFormat != "" && req.EncodingFormat != "float" && req.EncodingFormat != "base64" {
		s.writeError(w, http.StatusBadRequest, errInvalidRequest, "encoding_format must be float or base64")
		return
	}
	data := make([]oaiEmbedding, len(inputs))
	promptTokens := 0
	for i, in := range inputs {
		vec := s.embed(req.Model, in, dims)
		var emb any = vec
		if req.EncodingFormat == "base64" {
			emb = encodeFloat32Base64(vec)
		}
		data[i] = oaiEmbedding{Object: "embedding", Embedding: emb, Index: i}
		promptTokens += tokens.Count(in)
	}
	if !p.first() { // embeddings have no token stream: time to first token is the processing time
		return
	}
	w.Header().Set("X-Request-Id", newID("req_", 32))
	s.writeJSON(w, http.StatusOK, map[string]any{
		"object": "list", "data": data, "model": req.Model,
		"usage": map[string]int{"prompt_tokens": promptTokens, "total_tokens": promptTokens},
	})
}

// embeddingInputs accepts a string, an array of strings, or token arrays (hashed as text).
func embeddingInputs(raw json.RawMessage) ([]string, error) {
	var one string
	if json.Unmarshal(raw, &one) == nil {
		if one == "" {
			return nil, fmt.Errorf("'input' must not be empty")
		}
		return []string{one}, nil
	}
	var many []string
	if json.Unmarshal(raw, &many) == nil {
		if len(many) == 0 {
			return nil, fmt.Errorf("'input' must not be empty")
		}
		return many, nil
	}
	var tokenArrays [][]int
	if json.Unmarshal(raw, &tokenArrays) == nil && len(tokenArrays) > 0 {
		out := make([]string, len(tokenArrays))
		for i, t := range tokenArrays {
			out[i] = fmt.Sprint(t)
		}
		return out, nil
	}
	var tokenArray []int
	if json.Unmarshal(raw, &tokenArray) == nil && len(tokenArray) > 0 {
		return []string{fmt.Sprint(tokenArray)}, nil
	}
	return nil, fmt.Errorf("'input' must be a string, an array of strings or token arrays")
}

// maxEmbeddingDimensions bounds vector size (and the allocation per input).
const maxEmbeddingDimensions = 8192

// embed returns a deterministic unit vector for (seed, model, input).
func (s *Server) embed(model, input string, dims int) []float32 {
	if dims < 1 || dims > maxEmbeddingDimensions {
		dims = s.cfg.Embeddings.Dimensions
	}
	rng := s.seeds.content(requestHash("embedding:"+model, []byte(input)))
	vec := make([]float32, dims)
	var norm float64
	for i := range vec {
		v := rng.NormFloat64()
		vec[i] = float32(v)
		norm += v * v
	}
	norm = math.Sqrt(norm)
	for i := range vec {
		vec[i] = float32(float64(vec[i]) / norm)
	}
	return vec
}

func encodeFloat32Base64(vec []float32) string {
	buf := make([]byte, 4*len(vec))
	for i, v := range vec {
		binary.LittleEndian.PutUint32(buf[4*i:], math.Float32bits(v))
	}
	return base64.StdEncoding.EncodeToString(buf)
}

// modelCreated is a fixed timestamp so /v1/models is deterministic.
const modelCreated = 1721172741

func (s *Server) handleOpenAIModels(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.applyFaults(w, r, requestHash(r.URL.Path, nil)); !ok {
		return
	}
	data := make([]map[string]any, 0, len(s.cfg.Models))
	for _, m := range s.cfg.Models {
		data = append(data, map[string]any{"id": m, "object": "model", "created": modelCreated, "owned_by": "tollgate-sim"})
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"object": "list", "data": data})
}

func (s *Server) handleOpenAIModel(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.applyFaults(w, r, requestHash(r.URL.Path, nil)); !ok {
		return
	}
	m := r.PathValue("model")
	if !s.models[m] {
		s.writeError(w, http.StatusNotFound, errModelNotFound, fmt.Sprintf("The model `%s` does not exist or you do not have access to it.", m))
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]any{"id": m, "object": "model", "created": modelCreated, "owned_by": "tollgate-sim"})
}
