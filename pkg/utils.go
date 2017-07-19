package lib

import (
	"io/ioutil"
	"net/http"
)

func Check(e error) {
	if e != nil {
		panic(e)
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
