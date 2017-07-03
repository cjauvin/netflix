package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	daysBack = 20
	country  = "CA"
)

type apiResponse struct {
	Count string     `json:"COUNT"`
	Items [][]string `json:"ITEMS"`
}

type item struct {
	//itemID    int
	netflixID int
	imdbID    string
	title     string
	summary   string
	itemType  string
	year      int
	apiDate   time.Time
	duration  string
	imageUrl  string
	image     []byte
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func buildItem(values []string) (it *item, err error) {
	netflixID, err := strconv.Atoi(values[0])
	if err != nil {
		return
	}
	year, err := strconv.Atoi(values[7])
	if err != nil {
		return
	}
	apiDate, err := time.Parse("2006-01-02", values[10])
	if err != nil {
		return
	}
	img, err := downloadImage(values[2])
	if err != nil {
		return
	}
	it = &item{
		netflixID: netflixID,
		imdbID:    values[11],
		title:     values[1],
		summary:   values[3],
		itemType:  values[6],
		year:      year,
		apiDate:   apiDate,
		duration:  values[8],
		imageUrl:  values[2],
		image:     img,
	}
	return
}

func insert(db *sql.DB, it *item) (err error) {
	fmt.Println(it.title)
	_, err = db.Exec("insert into item (netflix_id, imdb_id, title, summary, item_type, year, api_date, duration, image_url, image) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", it.netflixID, it.imdbID, it.title, it.summary, it.itemType, it.year, it.apiDate, it.duration, it.imageUrl, it.image)
	return
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

func main() {
	db, err := sql.Open("postgres", "host=/var/run/postgresql dbname=netflix sslmode=disable")
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

		apiResp := apiResponse{}

		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		check(err)

		for _, values := range apiResp.Items {
			it, err := buildItem(values)
			check(err)
			err = insert(db, it)
			check(err)
		}

		done = true
	}

}
