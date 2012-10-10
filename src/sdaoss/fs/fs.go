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
}

type Bucket struct {
	EntryURI, Delimiter string
	Location string `xml:"LocationConstraint"`
	Objs []*Object `xml:"Contents"`
	CommonPrefixes []string `xml:"CommonPrefixes>Prefix"`
}

type Group struct {

}