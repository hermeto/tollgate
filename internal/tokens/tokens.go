// Package tokens approximates LLM token counts without tokenizer files.
//
// Method (documented in docs/simulator.md): text is split the way o200k_base pre-tokenizes it —
// letter runs (with one optional leading space or symbol), CJK runs, digit groups of up to three,
// symbol runs and whitespace runs — and each piece is converted to tokens with constants fitted
// against tiktoken o200k_base on testdata/tokens/corpus.jsonl. The simulator reports usage with this
// method and the gateway estimates with it, so both agree by construction; the error against the
// real tokenizer is measured by TestErrorAgainstReference.
package tokens

import (
	"unicode"
	"unicode/utf8"
)

// params are the fitted constants of the approximation.
type params struct {
	wordFree     int     // letters covered by the first token of an ASCII word
	wordStep     int     // extra letters per additional token of an ASCII word
	intlFree     int     // same, for words with non-ASCII letters
	intlStep     int     // same, for words with non-ASCII letters
	cjkPerToken  float64 // CJK characters per token
	symbolsPer   int     // symbol characters per token
	spacesPer    int     // extra whitespace characters per token (beyond the one absorbed by a word)
	newlineToken int     // tokens per newline run
}

// fitted was chosen by grid search minimising mean absolute percentage error against tiktoken
// o200k_base on testdata/tokens/corpus.jsonl (in-sample 7.4%, bias +2.2%; fitting on half the corpus
// and testing on the other half gives 5.8%). Regenerate the reference with scripts/tokens/reference.py.
var fitted = params{
	wordFree: 9, wordStep: 4, intlFree: 6, intlStep: 5,
	cjkPerToken: 1.5, symbolsPer: 3, spacesPer: 4, newlineToken: 1,
}

type class uint8

const (
	clsNone class = iota
	clsLetter
	clsCJK
	clsDigit
	clsSpace
	clsNewline
	clsSymbol
)

// asciiClass is a lookup table for the ASCII fast path.
var asciiClass = func() (t [utf8.RuneSelf]class) {
	for r := range rune(utf8.RuneSelf) {
		t[r] = classifySlow(r)
	}
	return t
}()

func classify(r rune) class {
	if r < utf8.RuneSelf {
		return asciiClass[r]
	}
	return classifySlow(r)
}

func classifySlow(r rune) class {
	switch {
	case r == '\n' || r == '\r':
		return clsNewline
	case unicode.IsSpace(r):
		return clsSpace
	case unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul):
		return clsCJK
	case unicode.IsLetter(r) || unicode.Is(unicode.M, r):
		return clsLetter
	case unicode.IsDigit(r):
		return clsDigit
	default:
		return clsSymbol
	}
}

// Count returns the approximate number of tokens in text.
func Count(text string) int {
	return countWith(text, fitted)
}

func countWith(text string, p params) int {
	n := 0
	i := 0
	for i < len(text) {
		r, size := rune(text[i]), 1
		if r >= utf8.RuneSelf {
			r, size = utf8.DecodeRuneInString(text[i:])
		}
		c := classify(r)
		j := i + size
		runes := 1
		ascii := r < utf8.RuneSelf
		for j < len(text) {
			if b := text[j]; b < utf8.RuneSelf {
				if asciiClass[b] != c {
					break
				}
				j++
				runes++
				continue
			}
			r2, s2 := utf8.DecodeRuneInString(text[j:])
			if classify(r2) != c {
				break
			}
			ascii = false
			j += s2
			runes++
		}
		switch c {
		case clsLetter:
			if ascii {
				n += wordTokens(runes, p.wordFree, p.wordStep)
			} else {
				n += wordTokens(runes, p.intlFree, p.intlStep)
			}
		case clsCJK:
			n += ceilDiv(float64(runes), p.cjkPerToken)
		case clsDigit:
			n += (runes + 2) / 3
		case clsSymbol:
			n += (runes + p.symbolsPer - 1) / p.symbolsPer
		case clsNewline:
			n += p.newlineToken
		case clsSpace:
			// A single space before a word is absorbed by the word's first token.
			extra := runes - 1
			if j >= len(text) || nextIsNewline(text[j:]) {
				extra = runes
			}
			if extra > 0 {
				n += (extra + p.spacesPer - 1) / p.spacesPer
			}
		}
		i = j
	}
	return n
}

func nextIsNewline(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return r == '\n' || r == '\r'
}

func wordTokens(runes, free, step int) int {
	if runes <= free {
		return 1
	}
	return 1 + (runes-free+step-1)/step
}

func ceilDiv(a, b float64) int {
	q := int(a / b)
	if float64(q)*b < a {
		q++
	}
	return q
}

// Message is the minimal chat message shape needed for counting.
type Message struct {
	Role    string
	Content string
}

// Per-message overhead of the OpenAI chat format (OpenAI cookbook, "How to count tokens").
const (
	tokensPerMessage = 3
	tokensForReply   = 3
)

// CountChat returns the approximate prompt tokens of a chat request: content tokens plus the
// per-message and reply-priming overhead of the chat format.
func CountChat(msgs []Message) int {
	n := tokensForReply
	for _, m := range msgs {
		n += tokensPerMessage + Count(m.Role) + Count(m.Content)
	}
	return n
}
