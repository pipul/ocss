package speedtest

import (
	"io/ioutil"
	"os"
	"fmt"
	"time"
	"runtime"
)


var (
	ps, nl string  // path and newline seperater
)

// Open cloud storage service
type ocss interface {
	Get(localfile, entryURI string) (code int, err error)
	Put(entryURI, mimeType, localfile string) (code int, err error)
	Delete(entryURI string) (code int, err error)
}

type storage struct {
	conn ocss
	sname, bucket string //service name and testing bucket
}

type Service struct {
	stg []storage
	count int
	debug bool
	datadir, resultdir string
}

func init() {
	ps = PathSep()
	nl = Newline()
	return
}


func New(count int, debug bool, datadir, resultdir string) *Service {
	return &Service{nil, count, debug, datadir, resultdir}
}

func GetTime() int64 {
	return time.Now().UnixNano() / 1000000 // ms, corresponse for Byte
}

func Tempfile(dir, prefix string) (fn string, err error) {
	f, err := ioutil.TempFile(dir, prefix)
	if err != nil {
		return
	}
	defer f.Close()
	fn = f.Name()
	return
}

func Newline() (nl string) {
	if runtime.GOOS == "windows" {
		nl = "\r\n"
	} else {
		nl = "\n"
	}
	return
}

func PathSep() (ps string) {
	if runtime.GOOS == "windows" {
		ps = "\\"
	} else {
		ps = "/"
	}
	return
}


func (s *Service) AddTestStorage(conn ocss, name, bucket string) {
	s.stg = append(s.stg, storage{conn, name, bucket})
}

func (s *Service) getfilelist() (filelist map[string]string) {
	d, err := os.Open(s.datadir)
	if err != nil {
		return
	}
	defer d.Close()
	fns, err := d.Readdirnames(-1)
	if err != nil {
		return
	}

	filelist = make(map[string]string)
	for _, fn := range fns {
		fi, err := os.Stat(s.datadir + ps + fn)
		if err != nil {
			return
		}
		if fi.IsDir() == true {
			continue
		}
		filelist[fn] = s.datadir + ps + fn
	}
	return
}


func (s *Service) doPutTesting(stg storage) (minsp, maxsp, avgsp, sr int) {
	var spsum, cursp, code int
	var t, fsize int64
	var entryURI string

	upfiles := s.getfilelist()
	for i := 0; i < s.count; i++ {
		for k,v := range upfiles {
			// !important
			if stg.sname == "qnbox" { 
				entryURI = stg.bucket + ":" + k
			} else {
				entryURI = "/" + stg.bucket + "/" + k
			}
			fi, err := os.Stat(v)
			if err != nil {
				continue
			}
			fsize = fi.Size()
			t = GetTime()
			code, err = stg.conn.Put(entryURI, "", v)
			t = GetTime() - t

			if err == nil && code/100 == 2 {
				cursp = int(fsize/t)
				if minsp == 0 && maxsp == 0 {
					minsp = cursp
					maxsp = cursp
				} else if  cursp < minsp {
					minsp = cursp
				} else if  cursp > maxsp {
					maxsp = cursp
				}
				if s.debug == true {
					fmt.Printf("%10s PUT => key: %-10s speed: %-10d succ: ok%s", stg.sname, k, cursp, nl)
				}
				sr++
				spsum += cursp
			}
		}
	}
	if sr == 0 {
		avgsp = 0
	} else {
		avgsp = spsum/sr
	}
	return
}


func (s *Service) doGetTesting(stg storage) (minsp, maxsp, avgsp, sr int) {

	var spsum, cursp, code int
	var entryURI string
	var t int64

	upfiles := s.getfilelist()
	for i := 0; i < s.count; i++ {
		for k,_ := range upfiles {
			// !important
			if stg.sname == "qnbox" { 
				entryURI = stg.bucket + ":" + k
			} else {
				entryURI = "/" + stg.bucket + "/" + k
			}

			f, err := Tempfile("", "ocss")
			if err != nil {
				continue
			}
			t = GetTime()
			code, err = stg.conn.Get(f, entryURI)
			t = GetTime() - t

			if code/100 == 2 && err == nil {
				fi, err := os.Stat(f)
				if err != nil {
					os.Remove(f)
					continue
				}
				cursp = int(fi.Size()/t)
				if minsp == 0 && maxsp == 0 {
					minsp = cursp
					maxsp = cursp
				} else if cursp < minsp {
					minsp = cursp
				} else if cursp > maxsp {
					maxsp = cursp
				}
				if s.debug == true {
					fmt.Printf("%10s GET => key: %-10s speed: %-10d succ: ok%s", stg.sname, k, cursp, nl)
				}
				sr++
				spsum += cursp
			}
			os.Remove(f)
		}
	}
	if sr == 0 {
		avgsp = 0
	} else {
		avgsp = spsum/sr
	}
	return
}

func (s *Service) doCleanTesting() {
	var (
		entryURI string
	)
	cleanfiles := s.getfilelist()
	for _,stg := range s.stg {
		for k,_ := range cleanfiles {
			if stg.sname == "qnbox" { 
				entryURI = stg.bucket + ":" + k
			} else {
				entryURI = "/" + stg.bucket + "/" + k
			}
			stg.conn.Delete(entryURI)
		}
	}
}


func (s *Service) RunTest() {
	resfile := s.resultdir + ps + fmt.Sprintf("%d",time.Now().Unix())
	f, err := os.Create(resfile)
	if err != nil {
		return
	}
	defer f.Close()

	// Testing Upload Speed
	fmt.Fprintf(f, "PUT Testing (k/s) ==> %s%10s %10s %10s %10s %10s%s",
		nl, "Name", "Min", "Max", "Avg", "Success", nl)
	for _,stg := range s.stg {
		minsp, maxsp, avgsp, sr := s.doPutTesting(stg)
		fmt.Fprintf(f, "%10s %10d %10d %10d %10d%s", stg.sname, minsp, maxsp, avgsp, sr, nl)
	}

	// Testing Download Speed
	fmt.Fprintf(f, "GET Testing (k/s) ==> %s%10s %10s %10s %10s %10s%s",
		nl, "Name", "Min", "Max", "Avg", "Success", nl)
	for _,stg := range s.stg {
		minsp, maxsp, avgsp, sr := s.doGetTesting(stg)
		fmt.Fprintf(f, "%10s %10d %10d %10d %10d%s", stg.sname, minsp, maxsp, avgsp, sr, nl)
	}

	// Clean test data
	s.doCleanTesting()
}