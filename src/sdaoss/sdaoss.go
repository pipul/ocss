package sdaoss

import (
	"io"
	"os"
	"fmt"
	"strings"
	"strconv"
	"net/http"
	"sdaoss/oauth"
	"sdaoss/rpc"
	"sdaoss/fs"
	"encoding/xml"
)

type Config struct {
	Host string `json:"HOST"`
	Access_key string `json:"ACCESS_KEY"`
	Access_secret string `json:"ACCESS_SECRET"`
}

type Service struct {
	host string
	conn rpc.Client
}

func New(c Config) *Service {
	key := c.Access_key
	secret := c.Access_secret
	host := c.Host
	t := oauth.NewTransport(key, secret, nil)
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}


func (s *Service) Get(localfile, entryURI string) (code int, err error) {
	f, err := os.OpenFile(localfile, os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	o := &fs.Object {
		EntryURI: entryURI,
		Body: f,
	}
	return s.GetObject(o)
}
func (s *Service) Put(entryURI, mimeType, localfile string) (code int, err error) {
	f, err := os.Open(localfile)
	if err != nil {
		return
	}
	fi, err := os.Stat(localfile)
	if err != nil {
		return
	}
	o := &fs.Object {
		EntryURI: entryURI,
		Type: mimeType,
		Size: fi.Size(),
		Body: f,
	}
	return s.PutObject(o)
}

func (s *Service) Delete(entryURI string) (code int, err error) {
	o := &fs.Object {
		EntryURI: entryURI,
	}
	return s.DelObject(o)
}


func (s *Service) PutObject(o *fs.Object) (code int, err error) {
	url := s.host + o.EntryURI
	return s.conn.CallRet(nil, "PUT", url, o.Type, o.Body, o.Size)
}


type MultiUpload struct {
	Bucket,Key,UploadId string
}

func (s *Service) UploadInit(o *fs.Object) (code int, err error) {
	var ret MultiUpload

	url := s.host + o.EntryURI + "?uploads"
	code, err = s.conn.CallRet(&ret, "POST", url)
	o.EntryURI += "?uploadId=" + ret.UploadId
	return
}

func (s *Service) UploadPart(o *fs.Object, so *fs.Object, pN int) (code int, err error) {
	var resp *http.Response

	url := fmt.Sprintf("%s&partNumber=%d", o.EntryURI, pN)
	if so.EntryURI != "" {
		// Copy from remote host
		addi := map[string]interface{} {
			"x-snda-copy-source": so.EntryURI,
		}
		return s.conn.CallRet(so, "PUT", url, addi)
	} else {
		// Upload from localhost
		if resp, err = s.conn.Call("PUT", url, so.Body, so.Size); err != nil {
			return
		}
		defer resp.Body.Close()
		code = resp.StatusCode
		if code/100 == 2 {
			so.ETag = resp.Header.Get("ETag")
		} else {
			// Error handler
		}
	}
	return
}



type part struct {
	XMLName xml.Name `xml:"Part"`
	PartNumber int
	ETag string
}

type makeObject struct {
	XMLName xml.Name `xml:"CompleteMultipartUpload"`
	Ps []part
}

func (s *Service) MakeObject(o *fs.Object, so []*fs.Object) (code int, err error) {
	var mo makeObject

	url := s.host + o.EntryURI
	mo.Ps = make([]part, len(so))
	for k,v := range so {
		mo.Ps[k].PartNumber = k
		mo.Ps[k].ETag = v.ETag
	}

	b, err := xml.Marshal(&mo)
	if err != nil {
		return
	}
	bd := strings.NewReader(string(b))
	bdlen := int64(len(b))
	return s.conn.CallRet(o, "POST", url, bd, bdlen)

	return
}


func (s *Service) GetObject(o *fs.Object) (code int, err error) {
	url := s.host + o.EntryURI
	resp, err := s.conn.Call("GET", url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 && resp.ContentLength != 0 {
		n, err := io.Copy(o.Body, resp.Body)
		if err != nil {
			o.Size = n
		}
		o.ETag = resp.Header.Get("ETag")
		o.Type = resp.Header.Get("Content-Type")
		o.Mtime = resp.Header.Get("Last-Modified")
	}
	return
}

func (s *Service) HeadObject(o *fs.Object) (code int, err error) {
	url := s.host + o.EntryURI
	resp, err := s.conn.Call("HEAD", url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 {
		o.ETag = resp.Header.Get("ETag")
		o.Type = resp.Header.Get("Content-Type")
		o.Mtime = resp.Header.Get("Last-Modified")
		o.Size,err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			o.Size = 0
		}
	}
	return
}

func (s *Service) DelObject(o *fs.Object) (code int, err error) {
	url := s.host + o.EntryURI
	return s.conn.CallRet(nil, "DELETE", url)
}

func (s *Service) CopyObject(do *fs.Object, so *fs.Object) (code int, err error) {
	url := s.host + do.EntryURI
	addi := map[string]interface{} {
		"x-snda-copy-source": so.EntryURI,
	}
	return s.conn.CallRet(nil, "PUT", url, addi)
}


type newBucketRet struct {
	XMLName xml.Name `xml:"CreateBucketConfiguration"`
	Location string `xml:"LocationConstraint"`
}

func (s *Service) NewBucket(b *fs.Bucket) (code int, err error) {

	url :=  s.host + b.EntryURI
	if strings.Contains(s.host, "huadong-1") == true {
		b.Location = "huadong-1"
	} else {
		b.Location = "huabei-1"
	}
	c := &newBucketRet { Location: b.Location }
	b1, err := xml.Marshal(c)
	if err != nil {
		return
	}
	bd := strings.NewReader(string(b1))
	return s.conn.CallRet(nil, "PUT", url, bd, int64(len(b1)))
}

func (s *Service) DropBucket(b *fs.Bucket) (code int, err error) {
	url := s.host + b.EntryURI
	return s.conn.CallRet(nil, "DELETE", url)
}


func (s *Service) ListBucket(b *fs.Bucket) (code int, err error) {
	var (
		BucketName, Prefix string
	)
	for _,v := range strings.Split(b.EntryURI, "/") {
		if len(v) == 0 {
			continue
		}
		if BucketName == "" {
			BucketName = v
			continue
		}
		Prefix += v + "/"
	}
	if b.Delimiter == "" {
		b.Delimiter = "/"
	}
	url := s.host + "/" + BucketName + "?delimiter=" + b.Delimiter + "&prefix=" + Prefix
	code, err = s.conn.CallRet(b, "GET", url)
	if code/100 == 2 {
		for _,o := range b.Objs {
			o.EntryURI = "/" + BucketName + "/" + o.EntryURI
		}
		for k,_ := range b.CommonPrefixes {
			b.CommonPrefixes[k] = "/" + BucketName + "/" + b.CommonPrefixes[k]
		}
	}
	return
}