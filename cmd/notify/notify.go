package main

import (
	"flag"
	"log"

	lib "github.com/cjauvin/netflix/pkg"
)

func main() {

	pw := flag.String("pw", "", "STMP password")
	flag.Parse()
	if *pw == "" {
		log.Fatalf("pw must be provided")
	}

	db, err := lib.GetNetflixDB()
	lib.Check(err)
	defer db.Close()

	users, err := db.GetUsers()
	lib.Check(err)

	for _, u := range users {

		items, err := db.GetItems(u.LastSentItemID)
		lib.Check(err)

		if len(items) == 0 {
			log.Printf("No items for %s, skipping", u.Email)
			continue
		}

		body := lib.BuildEmailBody(items)

		err = lib.SendEmail("cjauvin@gmail.com", u.Email, "Netflix Updates", body, *pw)
		lib.Check(err)

		lastSentItemID := items[len(items)-1].ItemID
		db.UpdateUserLastSentItemID(u.UserAccountID, lastSentItemID)

		log.Printf("Sent %d items to %s", len(items), u.Email)
	}
}
