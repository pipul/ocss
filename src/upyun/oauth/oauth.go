package oauth

import (
	"io"
	"fmt"
	"net/http"
	"crypto/md5"
)

type Transport struct {
	user,passwd string

	transport http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	h := md5.New()
	io.WriteString(h, t.passwd)
	passwdmd5 := fmt.Sprintf("%x", h.Sum(nil))
	h.Reset()

	stringtosign := req.Method + "&" + req.URL.Path + "&"
	stringtosign += req.Header.Get("Date") + "&"
	stringtosign += fmt.Sprintf("%d", req.ContentLength) + "&"
	stringtosign += passwdmd5

	io.WriteString(h, stringtosign)
	digest := fmt.Sprintf("%x", h.Sum(nil))
	token := t.user + ":" + digest
	req.Header.Set("Digest", stringtosign)
	req.Header.Set("Authorization", "UpYun " + token)
	return t.transport.RoundTrip(req)
}

func NewTransport(user, passwd string, t http.RoundTripper) * Transport {
	if t == nil {
		t = http.DefaultTransport
	}
	return &Transport{user, passwd, t}
}