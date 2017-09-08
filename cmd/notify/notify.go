package main

import (
	"flag"
	"fmt"
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

	minToSend := flag.Int("min", 10, "minimum new items to send email")
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

		nItems := len(items)

		if nItems < *minToSend {
			log.Printf("Not enough items (%d) for %s, skipping", nItems, u.Email)
			continue
		}

		body := lib.BuildEmailBody(items)

		err = lib.SendEmail("cjauvin@gmail.com", u.Email, fmt.Sprintf("Netflix Updates (%d)", nItems), body, *pw)
		lib.Check(err)

		lastSentItemID := items[nItems-1].ItemID
		tx.UpdateUserLastSentItemID(u.UserAccountID, lastSentItemID)

		log.Printf("Sent %d items to %s", nItems, u.Email)
	}
	tx.Commit()
}
