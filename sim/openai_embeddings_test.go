package sim

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math"
	"net/http"
	"testing"
)

func decodeBase64Floats(t *testing.T, s string) []float32 {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil || len(raw)%4 != 0 {
		t.Fatalf("bad base64 embedding: %v", err)
	}
	out := make([]float32, len(raw)/4)
	for i := range out {
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(raw[4*i:]))
	}
	return out
}

func TestOpenAIEmbeddings(t *testing.T) {
	ts := newOpenAI(t, 42)
	floatReq := `{"model":"text-embedding-3-small","input":["first text","second text"]}`
	status, body := doReq(t, http.MethodPost, ts.URL+"/v1/embeddings", floatReq, oaiAuth)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	assertSameShape(t, "embeddings_float.json", jsonShape(t, fixture(t, "openai/embeddings_float.json")), jsonShape(t, body))
	var floats struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &floats); err != nil || len(floats.Data) != 2 || len(floats.Data[0].Embedding) != 1536 {
		t.Fatalf("float embeddings: %v (%d)", err, len(floats.Data))
	}
	var norm float64
	for _, v := range floats.Data[0].Embedding {
		norm += float64(v) * float64(v)
	}
	if math.Abs(norm-1) > 1e-3 {
		t.Fatalf("vector norm² = %v, want 1", norm)
	}

	b64Req := `{"model":"text-embedding-3-small","input":"first text","encoding_format":"base64"}`
	_, body = doReq(t, http.MethodPost, ts.URL+"/v1/embeddings", b64Req, oaiAuth)
	assertSameShape(t, "embeddings_base64.json", jsonShape(t, fixture(t, "openai/embeddings_base64.json")), jsonShape(t, body))
	var b64 struct {
		Data []struct {
			Embedding string `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &b64); err != nil {
		t.Fatal(err)
	}
	decoded := decodeBase64Floats(t, b64.Data[0].Embedding)
	for i, v := range decoded {
		if v != floats.Data[0].Embedding[i] {
			t.Fatalf("base64 vector differs from float vector at %d", i)
		}
	}

	_, body = doReq(t, http.MethodPost, ts.URL+"/v1/embeddings", `{"model":"text-embedding-3-small","input":"x","dimensions":256}`, oaiAuth)
	if err := json.Unmarshal(body, &floats); err != nil || len(floats.Data[0].Embedding) != 256 {
		t.Fatalf("dimensions not honoured: %s", body)
	}
	for _, bad := range []string{`{"model":"text-embedding-3-small","input":""}`, `{"model":"text-embedding-3-small","input":[]}`,
		`{"model":"text-embedding-3-small","input":"x","dimensions":0}`, `{"model":"text-embedding-3-small","input":"x","encoding_format":"int8"}`} {
		if status, _ := doReq(t, http.MethodPost, ts.URL+"/v1/embeddings", bad, oaiAuth); status != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", bad, status)
		}
	}
}

func TestOpenAIModels(t *testing.T) {
	ts := newOpenAI(t, 42)
	_, body := doReq(t, http.MethodGet, ts.URL+"/v1/models", "", oaiAuth)
	assertSameShape(t, "models.json", jsonShape(t, fixture(t, "openai/models.json")), jsonShape(t, body))
	if status, _ := doReq(t, http.MethodGet, ts.URL+"/v1/models/gpt-4o-mini", "", oaiAuth); status != http.StatusOK {
		t.Fatalf("retrieve known model: %d", status)
	}
	status, body := doReq(t, http.MethodGet, ts.URL+"/v1/models/gpt-unknown", "", oaiAuth)
	if status != http.StatusNotFound {
		t.Fatalf("retrieve unknown model: %d", status)
	}
	assertSameShape(t, "error_404_model.json", jsonShape(t, fixture(t, "openai/error_404_model.json")), jsonShape(t, body))
}
