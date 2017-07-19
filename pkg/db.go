package lib

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

const (
	limitForFullQuery = 50
)

type NetflixDB struct {
	*sql.DB
}

type Item struct {
	ItemID    int
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

type User struct {
	UserAccountID  int
	Email          string
	IsActive       bool
	LastSentItemID sql.NullInt64
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

func (db *NetflixDB) UpsertUser(email string, isActive bool) (*User, error) {
	row := db.QueryRow("insert into user_account (email, is_active) values ($1, $2) on conflict (email) do update set is_active = $2 returning *", email, isActive)
	u := User{}
	err := row.Scan(&u.UserAccountID, &u.Email, &u.IsActive, &u.LastSentItemID)
	return &u, err
}

func (db *NetflixDB) UpdateUserLastSentItemID(userAccountID int, lastSentItemID int) (err error) {
	_, err = db.Exec("update user_account set last_sent_item_id = $1 where user_account_id = $2", lastSentItemID, userAccountID)
	return
}

func (db *NetflixDB) GetItems(minItemID sql.NullInt64) (items []*Item, err error) {
	var rows *sql.Rows
	if minItemID.Valid {
		rows, err = db.Query("select * from item where item_id > $1 order by item_id", minItemID.Int64)
	} else {
		rows, err = db.Query(`
		    with t as (
			select * from item order by item_id desc limit $1
		    )
		    select * from t order by item_id`, limitForFullQuery)
	}
	defer rows.Close()
	if err == nil {
		for rows.Next() {
			it := Item{}
			err := rows.Scan(&it.ItemID, &it.NetflixID, &it.ImdbID, &it.Title, &it.Summary, &it.ItemType, &it.Year, &it.ApiDate, &it.Duration, &it.ImageUrl, &it.Image)
			if err != nil {
				panic(err)
			}
			items = append(items, &it)
		}
	}
	return
}

func (db *NetflixDB) GetUsers() (users []*User, err error) {
	rows, err := db.Query("select * from user_account where is_active")
	defer rows.Close()
	if err == nil {
		for rows.Next() {
			u := User{}
			err := rows.Scan(&u.UserAccountID, &u.Email, &u.IsActive, &u.LastSentItemID)
			if err != nil {
				panic(err)
			}
			users = append(users, &u)
		}
	}
	return
}
