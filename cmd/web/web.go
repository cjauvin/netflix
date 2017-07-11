package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	nfdb "github.com/cjauvin/netflix/db"
)

var (
	pw   = flag.String("pw", "", "POST password")
	port = flag.Int("port", 80, "port")
)

func post(db nfdb.NetflixDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Not allowed!"))
			return
		}

		if r.Form["password"][0] != *pw {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized!"))
			return
		}

		email := r.Form["email"][0]
		isSubscribing := r.Form["action"][0] == "subscribe"

		// TODO: validate email address

		err := db.UpsertUser(email, isSubscribing)
		if err != nil {
			panic(err)
		}

		if isSubscribing {
			fmt.Fprintf(w, "%s has been subscribed", email)
		} else {
			fmt.Fprintf(w, "%s has been unsubscribed", email)
		}
	})
}

func main() {

	db, err := nfdb.GetNetflixDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	flag.Parse()
	if *pw == "" {
		log.Fatalf("POST password must be provided")
	}

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("/do", post(db))
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
