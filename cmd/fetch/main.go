package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	nfdb "github.com/cjauvin/netflix/db"

	"github.com/lib/pq"
)

const (
	daysBack                 = 20
	country                  = "CA"
	uniqueViolationErrorCode = "23505" // https://www.postgresql.org/docs/current/static/errcodes-appendix.html
)

type apiResponse struct {
	Count string     `json:"COUNT"`
	Items [][]string `json:"ITEMS"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	db, err := nfdb.GetNetflixDB()
	check(err)
	defer db.Close()
	k, err := ioutil.ReadFile("mashape_key.txt")
	check(err)
	key := strings.TrimSpace(string(k))

	done := false
	for page := 1; !done; page++ {
		//u := fmt.Sprintf("https://unogs-unogs-v1.p.mashape.com/api.cgi?q=get:new%d:%s&p=%d&t=ns&st=adv", daysBack, country, page)
		u := "http://localhost:8001/sample.json"
		req, err := http.NewRequest("GET", u, nil)
		check(err)
		req.Header.Set("X-Mashape-Key", key)
		req.Header.Set("Accept", "application/json")

		client := http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		resp, err := client.Do(req)
		check(err)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		check(err)

		apiResp := apiResponse{}

		//err = json.NewDecoder(resp.Body).Decode(&apiResp)
		err = json.Unmarshal(body, &apiResp)
		if err != nil {
			log.Fatalf("Got this response: %v", string(body))
		}

		for _, values := range apiResp.Items {
			it, err := nfdb.BuildItem(values)
			check(err)
			err = db.InsertItem(it)
			if err != nil {
				if err, ok := err.(*pq.Error); ok {
					if err.Code == uniqueViolationErrorCode {
						fmt.Println("already there")
						continue
					}
				}
				check(err)
			}
		}

		done = true
	}

}
