package main

import (
	"flag"
	"io"
	"log"
	"os"

	lib "github.com/cjauvin/netflix/pkg"
)

func main() {

	f := lib.LogFile("notify")
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	pw := flag.String("pw", "", "STMP password")
	flag.Parse()
	if *pw == "" {
		log.Fatalf("pw must be provided")
	}

	tx, err := lib.GetNetflixTx(nil)
	lib.Check(err)

	defer tx.DB.Close()

	users, err := tx.GetUsers()
	lib.Check(err)

	for _, u := range users {

		items, err := tx.GetItems(u.LastSentItemID)
		lib.Check(err)

		if len(items) == 0 {
			log.Printf("No items for %s, skipping", u.Email)
			continue
		}

		body := lib.BuildEmailBody(items)

		err = lib.SendEmail("cjauvin@gmail.com", u.Email, "Netflix Updates", body, *pw)
		lib.Check(err)

		lastSentItemID := items[len(items)-1].ItemID
		tx.UpdateUserLastSentItemID(u.UserAccountID, lastSentItemID)

		log.Printf("Sent %d items to %s", len(items), u.Email)
	}
	tx.Commit()
}
