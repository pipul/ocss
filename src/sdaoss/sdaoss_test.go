package sdaoss

import (
	"testing"
	"os"
	"io/ioutil"
	"encoding/json"
	"sdaoss/fs"
)

var ss *Service
var testkey, testbucket string

func init() {
	var conf Config
	testkey = "sdaoss.conf3"
	testbucket = "test_bucket_asdf"
	b, _ := ioutil.ReadFile("sdaoss.conf")
	json.Unmarshal(b, &conf)
	ss = New(conf)
}

func doTestPut(t *testing.T) {
	f, err := os.Open("sdaoss.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fi, _ := os.Stat("sdaoss.conf")
	o := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey,
		Size: fi.Size(),
		Body: f,
	}
	code, err := ss.PutObject(o)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestGet(t *testing.T) {

	EntryURI := "/" + testbucket + "/" + testkey
	code, err := ss.Get(EntryURI, "ooss")
	if code/100 != 2 {
		t.Fatal(err)
	}
/*
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
	code, err := ss.GetObject(o)
	if code/100 != 2 {
		t.Fatal(err)
	}
*/
}

func doTestDelete(t *testing.T) {
	o := &fs.Object {
		EntryURI: "/" + testbucket + "/" + testkey,
	}
	code, err := ss.DelObject(o)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestNewBucket(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: "/" + testbucket + "2",
	}
	code, err := ss.NewBucket(b)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestListBucket(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: "/" + testbucket,
	}
	code, err := ss.ListBucket(b)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
	t.Log(b)
	t.Log(b.Objs[0])
}

func doTestDropBucket(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: "/" + testbucket + "2",
	}
	code, err := ss.DropBucket(b)
	if code/100 != 2 {
		t.Fatal(err)
	}
}

func doTestListFolder(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: "/" + testbucket + "/book/data/",
	}
	code, err := ss.ListBucket(b)
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
	doTestNewBucket(t)
	doTestListBucket(t)
	doTestDropBucket(t)

	doTestListFolder(t)
}
