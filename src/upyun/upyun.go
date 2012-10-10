package upyun

import (
	"io"
	"os"
	"io/ioutil"
	"net/http"
	"strings"
	"strconv"
	"upyun/oauth"
	"upyun/rpc"
	"upyun/fs"
)

type Config struct {
	Host string `json:"HOST"`
	User string `json:"USER"`
	Passwd string `json:"PASSWORD"`
}

type Service struct {
	host string
	conn rpc.Client
}

func New(c Config) *Service {

	user := c.User
	passwd := c.Passwd
	host := c.Host

	t := oauth.NewTransport(user, passwd, nil)
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

	if o.Type == "" {
		o.Type = "application/octet-stream"
	}
	url := s.host + o.EntryURI
	addi := make(map[string]string)
	for k,v := range o.Meta {
		addi[k] = v
	}
	resp, err := s.conn.Call("POST", url, o.Type, o.Body, o.Size)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	if code/100 == 2 {
		// if the file is puted to image bucket
		for k,v := range resp.Header {
			if strings.HasPrefix(k, "x­upyun­") == false {
				continue
			}
			if o.Meta == nil {
				o.Meta = make(map[string]string)
			}
			o.Meta[k] = v[0]
		}
	}
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
	if code/100 == 2 && resp.ContentLength != 2 {
		o.Size, err = io.Copy(o.Body, resp.Body)
		if err != nil {
			o.Size = 0
		}
	}
	return
}


func (s *Service) DelObject(o *fs.Object) (code int, err error) {
	url := s.host + o.EntryURI
	resp, err := s.conn.Call("DELETE", url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	code = resp.StatusCode
	return
}


func (s *Service) ListBucket(b *fs.Bucket) (code int, err error) {
	url := s.host + b.EntryURI
	resp, err := s.conn.Call("GET", url)
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
		ret := strings.Split(string(data), "\n")
		b.Objs = make([]*fs.Object, len(ret))
		// filename type size modif
		for k,v := range ret {
			r := strings.Split(v, "\t")
			o := &fs.Object {
				EntryURI: b.EntryURI + r[0],
				Type: r[1],
				Mtime: r[3],
			}
			o.Size, err = strconv.ParseInt(r[2], 10, 64)
			if err != nil {
				o.Size = 0
			}
			b.Objs[k] = o
		}
	}
	return
}