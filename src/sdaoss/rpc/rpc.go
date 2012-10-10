package rpc

import (
	"io"
	"io/ioutil"
	"time"
	"strings"
	"net/http"
	"encoding/json"
	"encoding/xml"
)

type Client struct {
	*http.Client
}

func callRet(resp *http.Response, ret interface{}) (code int, err error) {
	var (
		b []byte
	)
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 {
		if ret == nil || resp.ContentLength == 0 {
			return
		}
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		ct := resp.Header.Get("Content-Type")
		if strings.Contains(ct, "application/xml") {
			err = xml.Unmarshal(b, ret)
		} else if strings.Contains(ct, "application/json") {
			err = json.Unmarshal(b, ret)
		}
	}
	return
}

func (r Client) Call(method, url string, args... interface{}) (resp *http.Response, err error) {

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
	req.ContentLength = bdlen

	for k,v := range addition {
		req.Header.Set(k, v.(string))
	}

	return r.Do(req)
}


func (r Client) CallRet(ret interface{}, method, url string, args... interface{}) (code int, err error) {
	resp, err := r.Call(method, url, args...)
	if err != nil {
		return
	}
	return callRet(resp, ret)
}