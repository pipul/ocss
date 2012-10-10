package upyun

import (
	"testing"
	"os"
	"io/ioutil"
	"encoding/json"
	"upyun/fs"
)

var us *Service
var testkey, testbucket string

func init() {
	var conf Config
	testkey = "fangdong"
	testbucket = "fangdongtestbucket"
	b, _ := ioutil.ReadFile("upyun.conf")
	json.Unmarshal(b, &conf)
	us = New(conf)
}

func doTestPut(t *testing.T) {
	f, err := os.Open("upyun.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fi, err := os.Stat("upyun.conf")
	if err != nil {
		t.Fatal(err)
	}
	o := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey,
		Size: fi.Size(),
		Body: f,
	}
	code, err := us.PutObject(o)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
}

func doTestGet(t *testing.T) {
	f, err := ioutil.TempFile("", "ocss")
	if err != nil {
		t.Fatal(err)
	}
	fn := f.Name()
	defer f.Close()
	defer os.Remove(fn)

	o := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey,
		Body: f,
	}
	code, err := us.GetObject(o)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
	t.Log(o)
}

func doTestDelete(t *testing.T) {
	o := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey,
	}
	code, err := us.DelObject(o)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
}

func doTestMkdir(t *testing.T) {
	o := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey + "/",
		Meta: map[string]string { "Folder": "create", },
	}
	code, err := us.PutObject(o)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
}

func doTestDeleteDir(t *testing.T) {
	o := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey + "/",
	}
	code, err := us.DelObject(o)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
}

func doTestListDir(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: "/" + testbucket + "/",
	}
	code, err := us.ListBucket(b)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
	t.Log(b)
	t.Log(b.Objs[1])
}

func TestDo(t *testing.T) {
	doTestPut(t)
	doTestGet(t)
	doTestDelete(t)
	doTestMkdir(t)
	doTestDeleteDir(t)
	doTestListDir(t)
}