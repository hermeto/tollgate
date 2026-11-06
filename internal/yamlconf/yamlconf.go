// Package yamlconf turns YAML configuration files into schema-validated JSON documents.
package yamlconf

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"go.yaml.in/yaml/v4"

	"github.com/hermeto/tollgate/schemas"
)

// ToJSON parses YAML, normalises map keys to strings, validates the document against
// schemas/<schemaName> and returns it as JSON. An empty document becomes {}.
func ToJSON(data []byte, schemaName string) ([]byte, error) {
	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	if raw == nil {
		raw = map[string]any{}
	}
	raw = Normalize(raw)
	if err := schemas.Validate(schemaName, raw); err != nil {
		return nil, err
	}
	return json.Marshal(raw)
}

// ReadFile reads and converts a YAML file; an empty path yields {} validated against the schema.
func ReadFile(path, schemaName string) ([]byte, error) {
	if path == "" {
		return ToJSON(nil, schemaName)
	}
	data, err := os.ReadFile(path) // #nosec G304 -- operator-supplied config path
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	out, err := ToJSON(data, schemaName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return out, nil
}

// Normalize converts YAML maps with non-string keys (e.g. 429: 0.5) into JSON-compatible maps.
func Normalize(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			t[k] = Normalize(val)
		}
		return t
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			switch kk := k.(type) {
			case string:
				m[kk] = Normalize(val)
			case int:
				m[strconv.Itoa(kk)] = Normalize(val)
			default:
				m[fmt.Sprint(kk)] = Normalize(val)
			}
		}
		return m
	case []any:
		for i := range t {
			t[i] = Normalize(t[i])
		}
		return t
	default:
		return v
	}
}
