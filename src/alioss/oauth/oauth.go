package oauth

import (
	"io"
	"strings"
	"net/http"
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
	io.WriteString(h, req.Header.Get("Content-Md5") + "\n")
	io.WriteString(h, req.Header.Get("Content-Type") + "\n")
	io.WriteString(h, req.Header.Get("Date") + "\n")

	xs := make([]string, 0)
	for k, _ := range req.Header {
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "x-oss-") == true {
			xs = append(xs, k)
		}
	}
	sort.Strings(xs)
	for _, k := range xs {
		lk := strings.ToLower(k)
		io.WriteString(h, lk + ":" + req.Header.Get(k) + "\n")
	}

	io.WriteString(h, req.URL.Path)

	digest := base64.StdEncoding.EncodeToString(h.Sum(nil))

	token := t.key + ":" + digest
	req.Header.Set("Authorization", "OSS " + token)
	return t.transport.RoundTrip(req)
}

func NewTransport(key, secret string, transport http.RoundTripper) *Transport {

	if transport == nil {
		transport = http.DefaultTransport
	}
	return &Transport{key, []byte(secret), transport}
}