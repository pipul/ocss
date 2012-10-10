package alioss

import (
	"io"
	"fmt"
	"os"
//	"io/ioutil"
	"strings"
	"strconv"
	"net/http"
	"alioss/oauth"
	"alioss/rpc"
	"alioss/fs"
//	"encoding/xml"
	"crypto/md5"
)

type Config struct {
	Host string `json:"HOST"`
	Access_key string `json:"ACCESS_ID"`
	Access_secret string `json:"ACCESS_KEY"`
}

type Service struct {
	Host string
	Conn rpc.Client
}

func New(c Config) *Service {
	key := c.Access_key
	secret := c.Access_secret
	host := c.Host
	t := oauth.NewTransport(key, secret, nil)
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

func (s *Service) Get(localfile string, entryURI string) (code int, err error) {
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
	var addi map[string]interface{}
	
	if o.Type == "" {
		o.Type = "application/octet-stream"
	}
	url := s.Host + o.EntryURI
	for k,v := range o.Meta {
		addi["x-oss-meta-" + k] = v
	}
	resp, err := s.Conn.Post("PUT", url, o.Type, o.Body, o.Size, addi)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 {
		o.ETag = resp.Header.Get("ETag")
	}
	return
}

func (s *Service) GetObject(o *fs.Object) (code int, err error) {
	url := s.Host + o.EntryURI
	resp, err := s.Conn.Post("GET", url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 && resp.ContentLength != 0 {
		o.Size, err = io.Copy(o.Body, resp.Body)
		if err != nil {
			return
		}
		o.ETag = resp.Header.Get("ETag")
		o.Type = resp.Header.Get("Content-Type")
		o.Mtime = resp.Header.Get("Last-Modified")
	}
	return
}


func (s *Service) CopyObject(do *fs.Object, so *fs.Object) (code int, err error) {
	url := s.Host + do.EntryURI
	addi := map[string]interface{} {
		"x-oss-copy-source": so.EntryURI,
	}
	if so.ETag != "" {
		addi["x-oss-copy-source-if-match"] = so.ETag
	}
	return s.Conn.CallRet(do, "PUT", url, addi)
/*
	resp, err := s.Conn.Post("PUT", url, addi)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 && resp.ContentLength != 0 {
		var b []byte
		if b, err = ioutil.ReadAll(resp.Body); err != nil {
			return
		}
		err = xml.Unmarshal(b, do)
	}
	return
*/
}

func (s *Service) HeadObject(o *fs.Object) (code int, err error) {
	url := s.Host + o.EntryURI
	resp, err := s.Conn.Post("HEAD", url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 {
		o.Type = resp.Header.Get("Content-Type")
		o.ETag = resp.Header.Get("ETag")
		o.Size, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			o.Size = 0
		}
		o.Mtime = resp.Header.Get("Last-Modified")
		if o.Meta == nil {
			o.Meta = make(map[string]string)
		}
		for k,v := range resp.Header {
			prefix := "x-oss-meta-"
			if strings.HasPrefix(k, prefix) == false {
				continue
			}
			o.Meta[strings.TrimLeft(k, prefix)] = v[0]
		}
	}
	return
}




func (s *Service) singleDelete(o *fs.Object) (code int, err error) {
	url := s.Host + o.EntryURI
	return s.Conn.CallRet(nil, "DELETE", url)
/*
	resp, err := s.Conn.Post("DELETE", url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	return
*/
}

func (s *Service) multiDelete(os []*fs.Object) (code int, err error) {

	h := md5.New()
	bucket := strings.Split(os[0].EntryURI, "/")[1]
	url := s.Host + "/" + bucket + "?delete"
	str := "<Delete><Quiet>true</Quiet>"
	for _,o := range os {
		str += "<Object>" + strings.TrimLeft(o.EntryURI, "/" + bucket + "/") + "</Object>"
	}
	str += "</Delete>"
	io.WriteString(h, str)
	bd := strings.NewReader(str)
	bdlen := int64(len(str))
	addi := map[string]interface{} {
		"Content-Md5": fmt.Sprintf("%x", h.Sum(nil)),
	}
	return s.Conn.CallRet(nil, "POST", url, bd, bdlen, addi)
/*
	resp, err := s.Conn.Post("POST", url, bd, bdlen, addi)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	return
*/
}


func (s *Service) DelObject(os... *fs.Object) (code int, err error) {
	cgr := make(map[string][]*fs.Object)

	for _,o := range os {
		bucket := strings.Split(o.EntryURI, "/")[0]
		if _,ok := cgr[bucket]; !ok {
			cgr[bucket] = make([]*fs.Object, 0)
		}
		cgr[bucket] = append(cgr[bucket], o)
	}

	for _,os_ := range cgr {
		if len(os_) != 1 {
			return s.multiDelete(os_)
		} else {
			return s.singleDelete(os_[0])
		}
	}
	return 200,nil
}


func (s *Service) NewBucket(b *fs.Bucket) (code int, err error) {
	url := s.Host + b.EntryURI
	return s.Conn.CallRet(nil, "PUT", url)
/*
	resp, err := s.Conn.Post("PUT", url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	return
*/
}

func (s *Service) DropBucket(b *fs.Bucket) (code int, err error) {
	url := s.Host + b.EntryURI
	return s.Conn.CallRet(nil, "DELETE", url)
/*
	resp, err := s.Conn.Post("DELETE", url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	return
*/
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
		} else {
			Prefix += v + "/"
		}
	}

	if b.Delimiter == "" {
		b.Delimiter = "/"
	}
	b.EntryURI = "/" + BucketName + "/" + Prefix
	url := s.Host + "/" + BucketName + "?delimiter=" + b.Delimiter + "&prefix=" + Prefix
	code, err = s.Conn.CallRet(b, "GET", url)
	if code/100 == 2 {
		for _,o := range b.Objs {
			o.EntryURI = "/" + BucketName + "/" + o.EntryURI
		}
		for k,_ := range b.CommonPrefixes {
			b.CommonPrefixes[k] = "/" + BucketName + "/" + b.CommonPrefixes[k] 
		}		
	}
	return
/*
	resp, err := s.Conn.Post("GET", url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 && resp.ContentLength != 0 {
		var data []byte
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}
		if err = xml.Unmarshal(data, b); err != nil {
			return
		}
		for _,o := range b.Objs {
			o.EntryURI = "/" + BucketName + "/" + o.EntryURI
		}
		for k,_ := range b.CommonPrefixes {
			b.CommonPrefixes[k] = "/" + BucketName + "/" + b.CommonPrefixes[k] 
		}
	}
	return
*/
}