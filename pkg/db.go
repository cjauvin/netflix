package lib

import (
	"database/sql"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

const (
	limitForFullQuery = 50
)

type NetflixTx struct {
	*sql.Tx
	DB *sql.DB
}

type Item struct {
	ItemID    int
	NetflixID int
	ImdbID    string
	Title     string
	Summary   string
	ItemType  string
	Year      int
	APIDate   time.Time
	Duration  string
	ImageURL  string
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
		APIDate:   apiDate,
		Duration:  values[8],
		ImageURL:  values[2],
		Image:     img,
	}
	return
}

func GetNetflixTx(db *sql.DB) (NetflixTx, error) {
	var err error
	if db == nil {
		db, err = sql.Open("postgres", "host=/var/run/postgresql dbname=netflix sslmode=disable")
		if err != nil {
			return NetflixTx{}, err
		}
	}
	tx, err := db.Begin()
	return NetflixTx{tx, db}, err
}

func (tx NetflixTx) InsertItem(it *Item) (err error) {
	_, err = tx.Exec("insert into item (netflix_id, imdb_id, Title, Summary, item_type, Year, api_date, Duration, image_url, Image) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", it.NetflixID, it.ImdbID, it.Title, it.Summary, it.ItemType, it.Year, it.APIDate, it.Duration, it.ImageURL, it.Image)
	return
}

func (tx NetflixTx) HasItem(netflixID int) (bool, error) {
	var exists bool
	err := tx.QueryRow("select exists (select 1 from item where netflix_id = $1)", netflixID).Scan(&exists)
	return exists, err
}

func (tx NetflixTx) UpsertUser(email string, isActive bool) (*User, error) {
	row := tx.QueryRow("insert into user_account (email, is_active) values ($1, $2) on conflict (email) do update set is_active = $2 returning *", email, isActive)
	u := User{}
	err := row.Scan(&u.UserAccountID, &u.Email, &u.IsActive, &u.LastSentItemID)
	return &u, err
}

func (tx NetflixTx) UpdateUserLastSentItemID(userAccountID int, lastSentItemID int) (err error) {
	_, err = tx.Exec("update user_account set last_sent_item_id = $1 where user_account_id = $2", lastSentItemID, userAccountID)
	return
}

func (tx NetflixTx) GetItems(minItemID sql.NullInt64) (items []*Item, err error) {
	var rows *sql.Rows
	if minItemID.Valid {
		rows, err = tx.Query("select * from item where item_id > $1 order by item_id", minItemID.Int64)
	} else {
		rows, err = tx.Query(`
		    with t as (
			select * from item order by item_id desc limit $1
		    )
		    select * from t order by item_id`, limitForFullQuery)
	}
	defer rows.Close()
	if err == nil {
		for rows.Next() {
			it := Item{}
			err := rows.Scan(&it.ItemID, &it.NetflixID, &it.ImdbID, &it.Title, &it.Summary, &it.ItemType, &it.Year, &it.APIDate, &it.Duration, &it.ImageURL, &it.Image)
			if err != nil {
				panic(err)
			}
			items = append(items, &it)
		}
	}
	return
}

func (tx NetflixTx) GetUsers() (users []*User, err error) {
	rows, err := tx.Query("select * from user_account where is_active")
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
