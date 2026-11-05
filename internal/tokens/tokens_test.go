package tokens

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestCountPieces(t *testing.T) {
	cases := []struct {
		text string
		want int
	}{
		{"", 0},
		{"a", 1},
		{"hello", 1},
		{"hello world", 2},
		{"internationalization", 4}, // 20 letters: 1 + ceil(11/4)
		{"1234567", 3},              // digit groups of up to three
		{"...", 1},
		{"a\n\nb", 3},
		{"中文测试", 3}, // ceil(4 / 1.5)
		{"trailing   ", 2},
	}
	for _, tc := range cases {
		if got := Count(tc.text); got != tc.want {
			t.Errorf("Count(%q) = %d, want %d", tc.text, got, tc.want)
		}
	}
}

func TestCountIsAdditiveAcrossPieces(t *testing.T) {
	a, b := "The gateway enforces budgets.", " Every team has one."
	if Count(a+b) != Count(a)+Count(b) {
		t.Fatalf("Count is not additive at a word boundary: %d != %d + %d", Count(a+b), Count(a), Count(b))
	}
}

func TestCountChatOverhead(t *testing.T) {
	msgs := []Message{{Role: "system", Content: "Be brief."}, {Role: "user", Content: "hi"}}
	want := 3 + (3 + Count("system") + Count("Be brief.")) + (3 + Count("user") + Count("hi"))
	if got := CountChat(msgs); got != want {
		t.Fatalf("CountChat = %d, want %d", got, want)
	}
	if CountChat(nil) != 3 {
		t.Fatalf("empty chat = %d, want 3 (reply priming)", CountChat(nil))
	}
}

// TestErrorAgainstReference enforces spec 002 SC-006: mean absolute error against tiktoken
// o200k_base on the committed corpus is at most 15% (and the aggregate bias at most 5%).
func TestErrorAgainstReference(t *testing.T) {
	f, err := os.Open("../../testdata/tokens/corpus.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	data, err := os.ReadFile("../../testdata/tokens/reference_o200k.json")
	if err != nil {
		t.Fatal(err)
	}
	var ref struct {
		Tokenizer string `json:"tokenizer"`
		Samples   []struct {
			ID     string `json:"id"`
			Tokens int    `json:"tokens"`
		} `json:"samples"`
	}
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Fatal(err)
	}
	if ref.Tokenizer != "o200k_base" {
		t.Fatalf("reference tokenizer = %q", ref.Tokenizer)
	}
	type row struct {
		id       string
		est, ref int
		relErr   float64
	}
	var rows []row
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for i := 0; sc.Scan(); i++ {
		var s struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(sc.Bytes(), &s); err != nil {
			t.Fatal(err)
		}
		if i >= len(ref.Samples) || ref.Samples[i].ID != s.ID {
			t.Fatalf("corpus and reference out of sync at %q; rerun scripts/tokens/reference.py", s.ID)
		}
		est := Count(s.Text)
		r := ref.Samples[i].Tokens
		rows = append(rows, row{s.ID, est, r, float64(est-r) / float64(r)})
	}
	if len(rows) != len(ref.Samples) {
		t.Fatalf("corpus has %d samples, reference %d", len(rows), len(ref.Samples))
	}
	var sumAbs, est, refTotal float64
	for _, r := range rows {
		sumAbs += math.Abs(r.relErr)
		est += float64(r.est)
		refTotal += float64(r.ref)
	}
	mape := sumAbs / float64(len(rows))
	bias := (est - refTotal) / refTotal
	sort.Slice(rows, func(i, j int) bool { return math.Abs(rows[i].relErr) > math.Abs(rows[j].relErr) })
	var worst strings.Builder
	for _, r := range rows[:5] {
		worst.WriteString(r.id + " ")
	}
	t.Logf("samples=%d mean_abs_error=%.1f%% bias=%+.1f%% worst=%s", len(rows), mape*100, bias*100, worst.String())
	if mape > 0.15 {
		t.Errorf("mean absolute error %.1f%% > 15%%", mape*100)
	}
	if math.Abs(bias) > 0.05 {
		t.Errorf("aggregate bias %+.1f%% outside ±5%%", bias*100)
	}
}

func BenchmarkCount4KB(b *testing.B) {
	text := strings.Repeat("The gateway enforces per-team budgets and rate limits on every request. ", 57)[:4096]
	b.SetBytes(int64(len(text)))
	b.ReportAllocs()
	for b.Loop() {
		_ = Count(text)
	}
}
