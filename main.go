package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

func isContentEncodingGzip(h http.Header) bool {
	vlist := h.Get("Content-Encoding")
	if string(vlist) == "gzip" {
		return true
	}
	return false
}

func main() {
	director := func(request *http.Request) {
		url := *request.URL
		url.Scheme = "http"
		url.Host = "localhost:9000"

		for h, vlist := range request.Header {
			log.Printf("%s: %s\n", h, vlist)
		}

		var b []byte = nil
		var err error
		if request.Body != nil {
			b, err = ioutil.ReadAll(request.Body)
			if isContentEncodingGzip(request.Header) {
				f, _ := gzip.NewReader(bytes.NewReader(b))
				io.Copy(os.Stdout, f)
			} else {
				io.Copy(os.Stdout, bytes.NewReader(b))
			}
			if err != nil {
				log.Fatal("%v", err)
			}
		} else {
			log.Printf("request body is nil")
		}

		req, err := http.NewRequest(request.Method, url.String(), bytes.NewBuffer(b))
		req.Header = request.Header
		*request = *req
		request.Body = ioutil.NopCloser(bytes.NewBuffer(b))
	}

	rp := &httputil.ReverseProxy{Director: director}
	server := http.Server{
		Addr:    ":9001",
		Handler: rp,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}
