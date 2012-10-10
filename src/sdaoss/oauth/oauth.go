package oauth

import (
	"io"
	"net/http"
	"strings"
	"sort"
	"encoding/base64"
	"crypto/sha1"
	"crypto/hmac"
)


type Transport struct {
	key string
	secret []byte

	transport http.RoundTripper
}


func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	h := hmac.New(sha1.New, t.secret)
	io.WriteString(h, req.Method + "\n")
	io.WriteString(h, req.Header.Get("Content-MD5") + "\n")
	io.WriteString(h, req.Header.Get("Content-Type") + "\n")
	io.WriteString(h, req.Header.Get("Date") + "\n")

	xs := make([]string, 0)
	for k, _ := range req.Header {
		if strings.HasPrefix(k, "x-snda-") == true {
			xs = append(xs, k)
		}
	}
	sort.Strings(xs)
	for _, v := range xs {
		io.WriteString(h, v + ": " + req.Header.Get(v) + "\n")
	}
	io.WriteString(h, req.URL.Path)

	digest := base64.StdEncoding.EncodeToString(h.Sum(nil))

	token := t.key + ":" + digest
	req.Header.Set("Authorization", "SNDA " + token)
	return t.transport.RoundTrip(req)
}


func NewTransport(key, secret string, t http.RoundTripper) *Transport {
	if t == nil {
		t = http.DefaultTransport
	}
	return &Transport{key, []byte(secret), t}
}
