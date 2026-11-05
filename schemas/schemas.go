// Package schemas embeds Tollgate's JSON Schemas and validates documents against them.
package schemas

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed *.schema.json
var files embed.FS

var (
	mu       sync.Mutex
	compiled = map[string]*jsonschema.Schema{}
)

// Schema returns the compiled schema stored in schemas/<name>.
func Schema(name string) (*jsonschema.Schema, error) {
	mu.Lock()
	defer mu.Unlock()
	if s, ok := compiled[name]; ok {
		return s, nil
	}
	data, err := files.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("schema %s: %w", name, err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("schema %s: %w", name, err)
	}
	c := jsonschema.NewCompiler()
	c.AssertFormat()
	if err := c.AddResource(name, doc); err != nil {
		return nil, fmt.Errorf("schema %s: %w", name, err)
	}
	s, err := c.Compile(name)
	if err != nil {
		return nil, fmt.Errorf("schema %s: %w", name, err)
	}
	compiled[name] = s
	return s, nil
}

// ValidateJSON validates a JSON document against schemas/<name>.
func ValidateJSON(name string, data []byte) error {
	s, err := Schema(name)
	if err != nil {
		return err
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if err := s.Validate(inst); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// Validate validates any value that encodes to JSON (for example a decoded YAML document).
func Validate(name string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("encode for validation: %w", err)
	}
	return ValidateJSON(name, data)
}
