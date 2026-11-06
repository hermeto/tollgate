package openai

import "testing"

func TestDecodeChatRetainsRaw(t *testing.T) {
	body := []byte(`{"model":"m","messages":[{"role":"user","content":"hi"}],"new_field":true}`)
	req, err := DecodeChat(body)
	if err != nil {
		t.Fatal(err)
	}
	if string(req.Raw) != string(body) {
		t.Fatal("raw body was not retained")
	}
}

func FuzzDecodeChat(f *testing.F) {
	f.Add([]byte(`{"model":"m","messages":[]}`))
	f.Fuzz(func(t *testing.T, body []byte) { _, _ = DecodeChat(body) })
}
