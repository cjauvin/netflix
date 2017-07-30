package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	lib "github.com/cjauvin/netflix/pkg"
)

const (
	daysBack = 7
	country  = "CA"
)

type apiResponse struct {
	Count string     `json:"COUNT"`
	Items [][]string `json:"ITEMS"`
}

func main() {

	f := lib.LogFile("fetch")
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	key := flag.String("key", "", "Mashape API key")
	flag.Parse()
	if *key == "" {
		log.Fatalln("key must be provided")
	}

	tx, err := lib.GetNetflixTx(nil)
	lib.Check(err)

	defer tx.DB.Close()

	done := false
	for page := 1; !done; page++ {
		u := fmt.Sprintf("https://unogs-unogs-v1.p.mashape.com/api.cgi?q=get:new%d:%s&p=%d&t=ns&st=adv", daysBack, country, page)
		// u := "http://localhost:8001/sample.json"
		req, err := http.NewRequest("GET", u, nil)
		lib.Check(err)
		req.Header.Set("X-Mashape-Key", *key)
		req.Header.Set("Accept", "application/json")

		client := http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		resp, err := client.Do(req)
		lib.Check(err)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		lib.Check(err)

		apiResp := apiResponse{}

		//err = json.NewDecoder(resp.Body).Decode(&apiResp)
		err = json.Unmarshal(body, &apiResp)
		if err != nil {
			log.Fatalf("Got this response: %v", string(body))
		}

		for _, values := range apiResp.Items {
			it, err := lib.BuildItem(values)
			lib.Check(err)
			present, err := tx.HasItem(it.NetflixID)
			lib.Check(err)
			if !present {
				log.Println(it.Title)
				err = tx.InsertItem(it)
				lib.Check(err)
			} else {
				log.Printf("%s: <SKIPPING>\n", it.Title)
			}
		}

		done = len(apiResp.Items) == 0
		// done = true
	}
	tx.Commit()
}
