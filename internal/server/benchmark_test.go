package server

import("io";"net/http";"net/http/httptest";"testing";"github.com/hermeto/tollgate/internal/config")
func BenchmarkPassthroughContract(b *testing.B){cfg:=config.Default();cfg.Upstream.BaseURL="http://127.0.0.1:1";s,err:=New(cfg);if err!=nil{b.Fatal(err)};req:=httptest.NewRequest(http.MethodGet,"/healthz",nil);for i:=0;i<b.N;i++{w:=httptest.NewRecorder();s.handleOps(w,req);_,_=io.Copy(io.Discard,w.Result().Body)}}
