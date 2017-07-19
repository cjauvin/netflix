package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"

	_ "github.com/wader/disable_sendfile_vbox_linux"

	lib "github.com/cjauvin/netflix/pkg"
)

var (
	webPassword  = flag.String("web-pw", "", "POST password")
	SMTPPassword = flag.String("smtp-pw", "", "SMTP password")
	port         = flag.Int("port", 80, "port")
)

func post(db lib.NetflixDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Not allowed!"))
			return
		}

		_, emailOK := r.Form["email"]
		_, passwordOK := r.Form["password"]
		_, actionOK := r.Form["action"]

		if !emailOK || !passwordOK || !actionOK {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request!"))
			return
		}

		if r.Form["password"][0] != *webPassword {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized!"))
			return
		}

		email := r.Form["email"][0]
		isSubscribing := r.Form["action"][0] == "subscribe"

		re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
		if !re.MatchString(email) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request!"))
			return
		}

		u, err := db.UpsertUser(email, isSubscribing)
		lib.Check(err)

		if isSubscribing {
			items, err := db.GetItems(u.LastSentItemID)
			lib.Check(err)
			if len(items) > 0 {
				body := lib.BuildEmailBody(items)
				err = lib.SendEmail("cjauvin@gmail.com", u.Email, "Netflix Updates", body, *SMTPPassword)
				lib.Check(err)
				lastSentItemID := items[len(items)-1].ItemID
				db.UpdateUserLastSentItemID(u.UserAccountID, lastSentItemID)
				log.Printf("Sent %d items to %s", len(items), u.Email)
				fmt.Fprintf(w, "%s has been subscribed (a first email has been sent)", email)
			} else {
				fmt.Fprintf(w, "%s has been subscribed", email)
			}
		} else {
			fmt.Fprintf(w, "%s has been unsubscribed", email)
		}
	})
}

func main() {

	db, err := lib.GetNetflixDB()
	lib.Check(err)
	defer db.Close()

	flag.Parse()
	if *webPassword == "" {
		log.Fatalf("POST password must be provided")
	}
	if *SMTPPassword == "" {
		log.Fatalf("SMTP password must be provided")
	}

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("/do", post(db))
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
