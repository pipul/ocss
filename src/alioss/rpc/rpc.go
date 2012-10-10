package rpc

import (
	"io"
	"io/ioutil"
	"time"
	"strings"
	"errors"
	"net/http"
	"encoding/json"
	"encoding/xml"
)

var (
	AccessDenied           = 403
	BucketAlreadyExists    = 409
	BucketNotEmpty         = 409
	EntityTooLarge         = 400
	EntityTooSmall         = 400
	FileGroupTooLarge      = 400
	FilePartNotExist       = 400
	FilePartStale          = 400
	InvalidArgument        = 400
	InvalidAccessKeyId     = 403
	InvalidBucketName      = 400
	InvalidDigest          = 400
	InvalidObjectName      = 400
	InvalidPart            = 400
	InvalidPartOrder       = 400
	InternalError          = 500
	MalformedXML           = 400
	MethodNotAllowed       = 405
	MissingArgument        = 411
	MissingContentLength   = 411
	NoSuchBucket           = 404
	NoSuchKey              = 404
	NoSuchUpload           = 404
	NotImplemented         = 501
	PreconditionFailed     = 412
	RequestTimeTooSkewed   = 403
	RequestTimeout         = 400
	SignatureDoesNotMatch  = 403
	TooManyBuckets         = 501
)

type Client struct {
	*http.Client
}

type errRet struct {
	Code, Message, RequestId, HostId string
}

func callRet(ret interface{}, resp *http.Response) (code int, err error) {
	var (
		b []byte
		er errRet
	)
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
		if strings.Contains(ct, "application/json") {
			err = json.Unmarshal(b, ret)
		} else { // default decoding
			err = xml.Unmarshal(b, ret)
		}
	} else {
		if resp.ContentLength == 0 {
			return
		}
		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		if err = xml.Unmarshal(b, &er); err != nil {
			return
		}
		err = errors.New(er.Code)
	}
	return
}



func (r Client) Post(method, url string, args ...interface{}) (resp *http.Response, err error) {

	var (
		bdtype string
		bd io.Reader
		bdlen int64
		addition map[string]interface{}
	)
	t := time.Now().UTC()

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

	req.Header.Set("Content-Type", bdtype)
	req.Header.Add("Date", t.Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	req.ContentLength = bdlen

	for k,v := range addition {
		req.Header.Set(k, v.(string))
	}

	return r.Do(req)
}



func (r Client) CallRet(ret interface{}, method, url string, args... interface{}) (code int, err error) {

	resp, err := r.Post(method, url, args...)
	if err != nil {
		return
	}
	return callRet(ret, resp)

}

