package rpc

import (
	"io"
	"time"
	"net/http"
)


type Client struct {
	*http.Client
}

func (r Client) Call(method, url string, args ...interface{}) (resp *http.Response, err error) {

	var (
		bdtype string
		bd io.Reader
		bdlen int64
		addition map[string]interface{}
	)
	for _,arg := range args {
		switch arg.(type) {
		case string:
			bdtype = arg.(string)
		case io.Reader:
			bd = arg.(io.Reader)
		case int64:
			bdlen = arg.(int64)
		case map[string]interface{}:
			addition = arg.(map[string]interface{})
		}
	}
	req, err := http.NewRequest(method, url, bd)
	if err != nil {
		return
	}
	if bdtype == "" {
		if bd != nil {
			bdtype = "application/octet-stream"
		}
	}
	req.Header.Set("Content-Type", bdtype)
	t := time.Now().UTC()
	req.Header.Set("Date", t.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	for k,v := range addition {
		req.Header.Set(k, v.(string))
	}
	req.ContentLength = bdlen
	return r.Do(req)
}
