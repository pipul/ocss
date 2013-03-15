package main

import (
	"io/ioutil"
	"errors"
	"upyun"
	"alioss"
	"sdaoss"
	"qnbox"
	"testing"
	"speedtest"
	"encoding/json"
)

var (
	t testing.T
	T *speedtest.Service

	us *upyun.Service
	ss *sdaoss.Service
	as *alioss.Service
	qs *qnbox.Service
)

func loadconfig(ret interface{}, file string) (err error) {
	if ret == nil {
		return errors.New("Nothing to do")
	}
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	return json.Unmarshal(b, ret)
}


func init() {
	var (
		err error
		upyunconf upyun.Config
		sdaossconf sdaoss.Config
		aliossconf alioss.Config
		qnboxconf qnbox.Config
	)


	// Create Testing
	T = speedtest.New(1, true, "data", ".")

	// Alioss
	if err = loadconfig(&aliossconf, "config/alioss.conf"); err != nil {
		t.Fatal(err)
	}
	as = alioss.New(aliossconf)
	T.AddTestStorage(as, "alioss", "test_bucket_asdf")

	// Shengda oss
	if err = loadconfig(&sdaossconf, "config/sdaoss.conf"); err != nil {
		t.Fatal(err)
	}
	ss = sdaoss.New(sdaossconf)
	T.AddTestStorage(ss, "sdaoss", "test_bucket_asdf")


	// YouPai oss
	if err = loadconfig(&upyunconf, "config/upyun.conf"); err != nil {
		t.Fatal(err)
	}
	us = upyun.New(upyunconf)
	T.AddTestStorage(us, "upyun", "fangdongtestbucket")



	// Qiniu oss
	if err = loadconfig(&qnboxconf, "config/qnbox.conf"); err != nil {
		t.Fatal(err)
	}
	qs = qnbox.New(qnboxconf)
	T.AddTestStorage(qs, "qnbox", "qnbox_bucket")
}

func main() {
	T.RunTest()
}haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
haha
