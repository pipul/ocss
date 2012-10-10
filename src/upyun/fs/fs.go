package fs

import (
	"io"
)

type Object struct {
	EntryURI string
	Type string
	Size int64
	Mtime string
	Body io.ReadWriter
	Meta map[string]string
}

type Bucket struct {
	EntryURI string
	Objs []*Object
}