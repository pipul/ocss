package alioss

import (
	"os"
	"io/ioutil"
	"testing"
	"strconv"
	"alioss/fs"
	"encoding/json"
)

var as *Service
var testbucket, testkey string

func init() {
	var conf Config
	testbucket = "test_bucket_asdf"
	testkey = "alioss.conf"
	b, _ := ioutil.ReadFile("alioss.conf")
	json.Unmarshal(b, &conf)
	as = New(conf)
}

func doTestPut(t *testing.T) {

	f, err := os.Open("alioss.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fi, _ := os.Stat("alioss.conf")
	o := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey,
		Size: fi.Size(),
		Type: "application/octet-stream",
		Body: f,
	}
	code, err := as.PutObject(o)
	if code/100 != 2 {
		t.Fatal(err)
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
	code, err := as.GetObject(o)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestDelete(t *testing.T) {

	o := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey,
	}
	code, err := as.DelObject(o)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestCopy(t *testing.T) {

	do := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey + "1",
	}
	so := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey,
	}
	code, err := as.CopyObject(do,so)
	if code/100 != 2 {
		t.Fatal(err)
	}
	code, err = as.DelObject(do)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestHead(t *testing.T) {

	o := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey,
	}
	code, err := as.HeadObject(o)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestMultiDelete(t *testing.T) {

	os := make([]*fs.Object, 0)
	so := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey,
	}
	for i := 0; i < 3; i++ {
		do := &fs.Object {
			EntryURI: "/" + testbucket + "/" + testkey + strconv.Itoa(i),
		}
		code, err := as.CopyObject(do,so)
		if code/100 != 2 {
			t.Fatal(err)
		}
		os = append(os, do)
	}
	code, err := as.DelObject(os[0], os[1], os[2])
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestNewBucket(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: "/test_bucket_2",
	}
	code, err := as.NewBucket(b)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
}

func doTestDropBucket(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: "/test_bucket_2",
	}
	code, err := as.DropBucket(b)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
}

func doTestListBucket(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: "/" + testbucket,
	}
	code, err := as.ListBucket(b)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
	t.Log(b)
	t.Log(b.Objs[0])
}

func doTestListFolder(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: "/" + testbucket + "/in",
	}
	code, err := as.ListBucket(b)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
	t.Log(b)
	t.Log(b.Objs[0])
}


func TestDo(t *testing.T) {
	doTestPut(t)
	doTestGet(t)
	doTestCopy(t)
//	doTestMultiDelete(t)
	doTestHead(t)
	doTestDelete(t)
	doTestNewBucket(t)
	doTestDropBucket(t)
	doTestListBucket(t)
	doTestListFolder(t)
}