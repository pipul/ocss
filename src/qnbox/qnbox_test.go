package qnbox

import (
	"os"
	"io/ioutil"
	"testing"
	"qnbox/fs"
	"encoding/json"
)


var qs *Service
var testbucket, testkey string

func init() {
	var conf Config
	testkey = "qnbox.conf"
	testbucket = "qnbox_bucket"
	b, err := ioutil.ReadFile("qnbox.conf")
	if err != nil {
		os.Exit(-1)
	}
	json.Unmarshal(b, &conf)
	qs = New(conf)
}

func doTestPut(t *testing.T) {
	f, err := os.Open("qnbox.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fi, _ := os.Stat("qnbox.conf")
	o := &fs.Object {
		EntryURI: testbucket + ":" + testkey,
		Body: f,
		Size: fi.Size(),
	}
	code, err := qs.PutObject(o)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
	t.Log(o)
}

func doTestGet(t *testing.T) {
	f, err := ioutil.TempFile("./", "ocss")
	if err != nil {
		t.Fatal(err)
	}
	fn := f.Name()
	defer f.Close()
	defer os.Remove(fn)

	o := &fs.Object {
		EntryURI: testbucket + ":" + testkey,
		Body: f,
	}
	code, err := qs.GetObject(o)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
	t.Log(o)
}

func doTestStat(t *testing.T) {
	o := &fs.Object {
		EntryURI: testbucket + ":" + testkey,
	}
	code, err := qs.StatObject(o)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
	t.Log(o)
}

func doTestMove(t *testing.T) {
	do := &fs.Object {
		EntryURI: testbucket + ":" + testkey + "1",
	}
	so := &fs.Object {
		EntryURI: testbucket + ":" + testkey,
	}
	code, err := qs.MoveObject(do, so)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
	code, err = qs.MoveObject(so, do)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
}

func doTestCopy(t *testing.T) {
	do := &fs.Object {
		EntryURI: testbucket + ":111",
	}
	so := &fs.Object {
		EntryURI: testbucket + ":" + testkey,
	}
	code, err := qs.CopyObject(do, so)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
/*
	code, err = qs.DelObject(do)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
*/
}

func doTestDel(t *testing.T) {
	o := &fs.Object {
		EntryURI: testbucket + ":" + testkey,
	}
	code, err := qs.DelObject(o)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
}


func doTestNewBucket(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: testbucket + "11",
	}
	code, err := qs.NewBucket(b)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
}


func doTestDropBucket(t *testing.T) {
	b := &fs.Bucket {
		EntryURI: testbucket + "11",
	}
	code, err := qs.DropBucket(b)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
}



func doTestRPut(t *testing.T) {

	f, err := os.Open("qnbox.conf")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fi, _ := os.Stat("qnbox.conf")
	o := &fs.Object {
		EntryURI: testbucket + ":" + testkey,
		Body: f,
		Size: fi.Size(),
	}
	code, err := qs.ResumablePut(o)
	if code/100 != 2 || err != nil {
		t.Fatal(code, err)
	}
	t.Log(o)

	no := &fs.Object {
		EntryURI: testbucket + ":" + testkey,
	}
	code, err = qs.StatObject(no)
	if code/100 != 2 {
		t.Fatal(code, err)
	}
	t.Log(no)
}



func TestDo(t *testing.T) {

	doTestPut(t)
	doTestGet(t)
	doTestStat(t)
	doTestMove(t)
	doTestCopy(t)
	doTestDel(t)
	doTestNewBucket(t)
	doTestDropBucket(t)
	doTestRPut(t)
}