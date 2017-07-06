package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type NetflixDB struct {
	*sql.DB
}

type Item struct {
	NetflixID int
	ImdbID    string
	Title     string
	Summary   string
	ItemType  string
	Year      int
	ApiDate   time.Time
	Duration  string
	ImageUrl  string
	Image     []byte
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

func BuildItem(values []string) (it *Item, err error) {
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
	it = &Item{
		NetflixID: netflixID,
		ImdbID:    values[11],
		Title:     values[1],
		Summary:   values[3],
		ItemType:  values[6],
		Year:      year,
		ApiDate:   apiDate,
		Duration:  values[8],
		ImageUrl:  values[2],
		Image:     img,
	}
	return
}

func GetNetflixDB() (NetflixDB, error) {
	db, err := sql.Open("postgres", "host=/var/run/postgresql dbname=netflix sslmode=disable")
	return NetflixDB{db}, err
}

func (db *NetflixDB) InsertItem(it *Item) (err error) {
	fmt.Println(it.Title)
	_, err = db.Exec("insert into item (netflix_id, imdb_id, Title, Summary, item_type, Year, api_date, Duration, image_url, Image) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", it.NetflixID, it.ImdbID, it.Title, it.Summary, it.ItemType, it.Year, it.ApiDate, it.Duration, it.ImageUrl, it.Image)
	return
}
