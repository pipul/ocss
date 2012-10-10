package fs

import (
	"io"
)


type Object struct {
	EntryURI string `xml:"Key"`
	Type string
	ETag string
	Mtime string `xml:"LastModified"`
	Size int64
	Body io.ReadWriter
	Meta map[string]string
}

type Bucket struct {
	EntryURI, Delimiter string
	Objs []*Object `xml:"Contents"`
	CommonPrefixes []string `xml:"CommonPrefixes>Prefix"`
}