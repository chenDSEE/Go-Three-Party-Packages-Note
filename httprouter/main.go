package main

import (
	"fmt"
	"main/httprouter"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	for _, pair := range p {
		// 多个 Param 的情况先，按着 URL 里面的顺序进行排列的
		fmt.Printf("[%s] --> [%s]\n", pair.Key, pair.Value)
	}
	fmt.Fprintf(w, "new connection with params[%s, %s]!\n", p.ByName("Params"), p.ByName("id"))
}

// 只需要改一下注册 SDKHandler 就可以直接在这里用 Params 了
func SDKHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context()) // 从 r.Context() 中取出，然后断言
	fmt.Fprintf(w, "SDKHandler get params[%s]!\n", params.ByName("Params"))
}

func main() {
	mux := httprouter.New()
	// curl -i http://localhost:8080/hello/aimer/post-id/123456
	mux.GET("/hello/:Params/post-id/:id", helloHandler)

	// http.Handler 兼容
	mux.HandlerFunc("GET", "/adapter/:Params", SDKHandler)

	server := http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}
	server.ListenAndServe()
}