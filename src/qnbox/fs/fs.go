package fs

import (
	"os"
)


type Block struct {
	Ctx string
	Offset int64
	RestSize int64
	Checksum string
}

type Object struct {
	EntryURI string
	URL string         `json:"url"`
	Type string        `json:"mimeType"`
	Size int64         `json:"fsize"`
	Hash string        `json:"hash"`
	Ctime int64        `json:"putTime"`
	Expiry int         `json:"expires"`
	Progress []*Block   // Resumable blockPut Progress
	Body *os.File
}

type Bucket struct {
	EntryURI string
	Objs []*Object
}