package qnbox

import (
	"io"
	"os"
	"errors"
	"strconv"
//	"log"
	"sync"
	"net/http"
	"encoding/base64"
	"hash/crc32"
	"qnbox/rpc"
	"qnbox/oauth"
	"qnbox/fs"
	"qnbox/utils/bytes"
)


const (
	InvalidCtx = 701 // UP: 无效的上下文(bput)，可能情况：Ctx非法或者已经被淘汰（太久未使用）
)

func EncodeURI(uri string) string {
	return base64.URLEncoding.EncodeToString([]byte(uri))
}


type Config struct {
	Host map[string]string `json:"HOST"`
	Access_key string `json:"ACCESS_KEY"`
	Access_secret string `json:"SECRET_KEY"`

	// Use for resumable put
	BlockBits uint `json:"BLOCK_BITS"`
	ChunkSize int64 `json:"PUT_CHUNK_SIZE"`
	PutRetryTimes int `json:"PUT_RETRY_TIMES"`
	ExpiresTime int `json:"EXPIRES_TIME"`
}

type Service struct {
	Config
	Conn rpc.Client
}

func New(c Config) *Service {
	key := c.Access_key
	secret := c.Access_secret
	t := oauth.NewTransport(key, secret, nil)
	client := &http.Client{Transport: t}
	return &Service{c, rpc.Client{client}}
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
	/* Single mode upload */
//	return s.PutObject(o)

	/* Resumable mode upload */
	return s.ResumablePut(o)
}

func (s *Service) Delete(entryURI string) (code int, err error) {
	o := &fs.Object {
		EntryURI: entryURI,
	}
	return s.DelObject(o)
}

func (s *Service) PutObject(o *fs.Object) (code int, err error) {
	url := s.Host["io"] + "/rs-put/" + EncodeURI(o.EntryURI)
	if o.Type != "" {
		url += "/mimeType/" + EncodeURI(o.Type)
	}
	code, err = s.Conn.CallWith64(o, url, "application/octet-stream", o.Body, o.Size)
	return
}

func (s *Service) GetObject(o *fs.Object) (code int, err error) {
	url := s.Host["rs"] + "/get/" + EncodeURI(o.EntryURI)
	if o.Hash != "" {
		url += "/base/" + o.Hash
	}
	if o.Expiry > 0 {
		url += "/expires/" + strconv.Itoa(o.Expiry)
	}
	code, err = s.Conn.Call(o, url)
	if code/100 == 2 && o.Body != nil {
		code, err = s.Conn.Download(o.Body, o.URL)
	}
	return
}

func (s *Service) StatObject(o *fs.Object) (code int, err error) {
	url := s.Host["rs"] + "/stat/" + EncodeURI(o.EntryURI)
	code, err = s.Conn.Call(o, url)
	return
}

func (s *Service) DelObject(o *fs.Object) (code int, err error) {
	url := s.Host["rs"] + "/delete/" + EncodeURI(o.EntryURI)
	return s.Conn.Call(nil, url)
}

func (s *Service) MoveObject(do *fs.Object, so *fs.Object) (code int, err error) {
	url := s.Host["rs"] + "/move/" + EncodeURI(so.EntryURI) + "/" + EncodeURI(do.EntryURI)
	return s.Conn.Call(nil, url)
}

func (s *Service) CopyObject(do *fs.Object, so *fs.Object) (code int, err error) {
	url := s.Host["rs"] + "/copy/" + EncodeURI(so.EntryURI) + "/" + EncodeURI(do.EntryURI)
	return s.Conn.Call(nil, url)
}

func (s *Service) NewBucket(b *fs.Bucket) (code int, err error) {
	url := s.Host["rs"] + "/mkbucket/" + EncodeURI(b.EntryURI)
	return s.Conn.Call(nil, url)
}

func (s *Service) DropBucket(b *fs.Bucket) (code int, err error) {
	url := s.Host["rs"] + "/drop/" + EncodeURI(b.EntryURI)
	return s.Conn.Call(nil, url)
}

func (s *Service) ListBucket(b *fs.Bucket) (code int, err error) {
	return
}



// ---------

type putRet struct {
	Ctx string  `json:"ctx"`
	Checksum string `json:"checksum"`
	Crc32 uint32 `json:"crc32"`
	Offset uint32 `json:"offset"`
}

func (s *Service) putBlock(o *fs.Object, Idx int) (code int, err error) {
	var (
		ret putRet
		url string
	)
	h := crc32.NewIEEE()
	prog := o.Progress[Idx]
	offbase := int64(Idx << s.BlockBits)

	initProg := func(p *fs.Block) {
		if Idx == len(o.Progress) - 1 {
			p.RestSize = o.Size - offbase
		} else {
			p.RestSize = 1 << s.BlockBits
		}
		p.Offset = 0
		p.Ctx = ""
		p.Checksum = ""
	}

	if prog.Ctx == "" {
		initProg(prog)
	}

	for prog.RestSize > 0 {
		bdlen := s.ChunkSize
		if bdlen > prog.RestSize {
			bdlen = prog.RestSize
		}
		retry := s.PutRetryTimes
	lzRetry:
		h.Reset()
		bd1 := io.NewSectionReader(o.Body, int64(offbase + prog.Offset), int64(bdlen))
		bd := io.TeeReader(bd1, h)
		if prog.Ctx == "" {
			url = s.Host["up"] + "/mkblk/" + strconv.FormatInt(prog.RestSize, 10)
		} else {
			url = s.Host["up"] + "/bput/" + prog.Ctx + "/" + strconv.FormatInt(prog.Offset, 10)
		}
		code, err = s.Conn.CallWith(&ret, url, "application/octet-stream", bd, int(bdlen))
		if err == nil {
			if ret.Crc32 == h.Sum32() {
				prog.Ctx = ret.Ctx
				prog.Offset += bdlen
				prog.RestSize -= bdlen
				continue
			} else {
				err = errors.New("ResumableBlockPut: Invalid Checksum")
			}
		}
		if code == InvalidCtx {
			initProg(prog)
			continue   // retry upload current block
		}
		if retry > 0 {
			retry--
			goto lzRetry
		}
		break
	}
	return
}

func (s *Service) mkObject(o *fs.Object) (code int, err error) {
	var (
		ctx string
	)
	for k,p := range o.Progress {
		if k == len(o.Progress) - 1 {
			ctx += p.Ctx
		} else {
			ctx += p.Ctx + ","
		}
	}
	bd := []byte(ctx)
	url := s.Host["up"] + "/rs-mkfile/" + EncodeURI(o.EntryURI) + "/fsize/" + strconv.FormatInt(o.Size, 10)
	code, err = s.Conn.CallWith(nil, url, "", bytes.NewReader(bd), len(bd))
	return
}


func (s *Service) ResumablePut(o *fs.Object) (code int, err error) {
	var (
		wg sync.WaitGroup
		failed bool
	)
	blockcnt := int((o.Size + (1 << s.BlockBits) - 1) >> s.BlockBits)
	o.Progress = make([]*fs.Block, blockcnt)
	wg.Add(blockcnt)

	for i := 0; i < blockcnt; i++ {
		o.Progress[i] = &fs.Block{}
		blkIdx := i
		task := func() {
			defer wg.Done()
			code, err = s.putBlock(o, blkIdx)
			if err != nil {
				failed = true
			}
		}
		go task()
	}
	wg.Wait()

	if failed {
		return 400, errors.New("ResumableBlockPut haven't done")
	}
	return s.mkObject(o)
}