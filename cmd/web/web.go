package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"

	_ "github.com/wader/disable_sendfile_vbox_linux"

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

		_, emailOK := r.Form["email"]
		_, passwordOK := r.Form["password"]
		_, actionOK := r.Form["action"]

		if !emailOK || !passwordOK || !actionOK {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request!"))
			return
		}

		if r.Form["password"][0] != *pw {
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
