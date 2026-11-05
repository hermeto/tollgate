package schemas

import (
	"io/fs"
	"strings"
	"testing"
)

func TestAllSchemasCompile(t *testing.T) {
	names, err := fs.Glob(files, "*.schema.json")
	if err != nil || len(names) == 0 {
		t.Fatalf("glob: %v %v", names, err)
	}
	for _, n := range names {
		if _, err := Schema(n); err != nil {
			t.Errorf("%s: %v", n, err)
		}
	}
}

func TestSimSchema(t *testing.T) {
	cases := []struct {
		name    string
		doc     string
		wantErr string
	}{
		{"minimal", `{"flavor":"openai"}`, ""},
		{"full", `{"flavor":"anthropic","port":9002,"seed":7,"latency":{"ttft":{"kind":"lognormal","mu":5,"sigma":0.4}},
			"output_tokens":{"kind":"normal","mean":100,"stddev":10,"min":1,"max":200},
			"faults":{"error_rates":{"429":0.1,"529":0.05},"retry_after_seconds":2,
			"stall_mid_stream":{"rate":0.1,"after_tokens":3,"duration":"2s"},"outage":{"start_after":"5s","duration":"30s","status":503}}}`, ""},
		{"unknown field", `{"flavor":"openai","bogus":1}`, "additional properties"},
		{"bad flavor", `{"flavor":"azure"}`, "flavor"},
		{"fixed needs value", `{"latency":{"ttft":{"kind":"fixed"}}}`, "value"},
		{"bad status key", `{"faults":{"error_rates":{"418":0.5}}}`, "418"},
		{"rate above one", `{"faults":{"reset_rate":1.5}}`, "maximum"},
		{"bad duration", `{"faults":{"outage":{"duration":"thirty seconds"}}}`, "duration"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateJSON("sim.schema.json", []byte(tc.doc))
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}
