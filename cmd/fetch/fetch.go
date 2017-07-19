package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	lib "github.com/cjauvin/netflix/pkg"

	"github.com/lib/pq"
)

const (
	daysBack                 = 7
	country                  = "CA"
	uniqueViolationErrorCode = "23505" // https://www.postgresql.org/docs/current/static/errcodes-appendix.html
)

type apiResponse struct {
	Count string     `json:"COUNT"`
	Items [][]string `json:"ITEMS"`
}

func main() {

	key := flag.String("key", "", "Mashape API key")
	flag.Parse()
	if *key == "" {
		log.Fatalf("key must be provided")
	}

	db, err := lib.GetNetflixDB()
	lib.Check(err)
	defer db.Close()

	done := false
	for page := 1; !done; page++ {
		u := fmt.Sprintf("https://unogs-unogs-v1.p.mashape.com/api.cgi?q=get:new%d:%s&p=%d&t=ns&st=adv", daysBack, country, page)
		//u := "http://localhost:8001/sample.json"
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
			err = db.InsertItem(it)
			if err != nil {
				if err, ok := err.(*pq.Error); ok {
					if err.Code == uniqueViolationErrorCode {
						fmt.Println("already there")
						continue
					}
				}
				lib.Check(err)
			}
		}

		done = len(apiResp.Items) == 0
	}

}
