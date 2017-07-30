package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func Check(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}

func downloadImage(url string) (img []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	img, err = ioutil.ReadAll(resp.Body)
	return
}

func LogFile(name string) *os.File {
	ex, err := os.Executable()
	Check(err)
	logPath := fmt.Sprintf("%s/%s.log", filepath.Dir(ex), name)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	Check(err)
	return f
}
