package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
)

func main() {
	learn()
	fmt.Println("Process Browser history")
	readChromeHistory()
	fmt.Println("Start Proxy at localhost:8080...")
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if isBadURL(r.URL.String()) {
			return r, goproxy.NewResponse(r,
				goproxy.ContentTypeText, http.StatusForbidden,
				r.URL.String()+" is a malicious site!")
		}
		return r, nil
	})

	log.Fatalln(http.ListenAndServe(":8080", proxy))
	session.Close()
}
