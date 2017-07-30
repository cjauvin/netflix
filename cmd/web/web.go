package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"

	// _ "github.com/wader/disable_sendfile_vbox_linux"

	lib "github.com/cjauvin/netflix/pkg"
)

var (
	webPassword  = flag.String("web-pw", "", "POST password")
	SMTPPassword = flag.String("smtp-pw", "", "SMTP password")
	port         = flag.Int("port", 80, "port")
	dumpRequests = flag.Bool("dump", false, "dump requests")
)

func post(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if *dumpRequests {
			rd, err := httputil.DumpRequest(r, false)
			lib.Check(err)
			log.Println(string(rd))
		}

		r.ParseForm()
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Not allowed!"))
			log.Println(http.StatusMethodNotAllowed)
			return
		}

		_, emailOK := r.Form["email"]
		_, passwordOK := r.Form["password"]
		_, actionOK := r.Form["action"]

		if !emailOK || !passwordOK || !actionOK {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request!"))
			log.Println(http.StatusBadRequest)
			return
		}

		if r.Form["password"][0] != *webPassword {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized!"))
			log.Println(http.StatusUnauthorized)
			return
		}

		email := r.Form["email"][0]
		isSubscribing := r.Form["action"][0] == "subscribe"

		re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
		if !re.MatchString(email) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request!"))
			log.Println(http.StatusBadRequest)
			return
		}

		tx, err := lib.GetNetflixTx(db)
		lib.Check(err)

		u, err := tx.UpsertUser(email, isSubscribing)
		lib.Check(err)

		tx.Commit()

		if isSubscribing {
			msg := fmt.Sprintf("%s has been subscribed", email)
			fmt.Fprint(w, msg)
			log.Println(msg)
			// Send first email if there are items to send
			go func() {
				tx, err := lib.GetNetflixTx(db)
				lib.Check(err)
				items, err := tx.GetItems(u.LastSentItemID)
				lib.Check(err)
				if len(items) == 0 {
					return
				}
				body := lib.BuildEmailBody(items)
				err = lib.SendEmail("cjauvin@gmail.com", u.Email, "Netflix Updates", body, *SMTPPassword)
				lib.Check(err)
				lastSentItemID := items[len(items)-1].ItemID
				tx.UpdateUserLastSentItemID(u.UserAccountID, lastSentItemID)
				log.Printf("Sent %d items to %s", len(items), u.Email)
				//fmt.Fprintf(w, "%s has been subscribed (a first email has been sent)", email)
				tx.Commit()
			}()
		} else {
			msg := fmt.Sprintf("%s has been unsubscribed", email)
			fmt.Fprint(w, msg)
			log.Println(msg)
		}
	})
}

func main() {

	f := lib.LogFile("web")
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)

	flag.Parse()
	if *webPassword == "" {
		log.Fatalf("POST password must be provided")
	}
	if *SMTPPassword == "" {
		log.Fatalf("SMTP password must be provided")
	}

	db, err := sql.Open("postgres", "host=/var/run/postgresql dbname=netflix sslmode=disable")
	lib.Check(err)

	defer db.Close()

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("/do", post(db))
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
