package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/smtp"

	nfdb "github.com/cjauvin/netflix/db"
)

const (
	nCols = 5
)

type row [nCols]nfdb.Item

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func sendEmail(from string, to string, subject string, body string, pw string) (err error) {

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	fr := fmt.Sprintf("From: %s\n", from)
	t := fmt.Sprintf("To: %s\n", to)
	sub := fmt.Sprintf("Subject: %s\n", subject)
	msg := []byte(fr + t + sub + mime + body)

	auth := smtp.PlainAuth("", from, pw, "smtp.gmail.com")
	err = smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, msg)
	return
}

func main() {

	pw := flag.String("pw", "", "STMP password")
	flag.Parse()
	if *pw == "" {
		log.Fatalf("pw must be provided")
	}

	db, err := nfdb.GetNetflixDB()
	check(err)
	defer db.Close()

	tmpl := template.Must(template.New("example").Parse(`
	<table>
	{{range .}}
	  <tr>
	    {{range .}}<td><img src={{.ImageUrl}}><img><br>{{.Title}}<hr>{{.Summary}}</td>{{end}}
	  </tr>
	{{end}}
	</table>
	`))

	g := []row{}

	items, err := db.GetItems()
	check(err)

	var j int
	for i, it := range items {
		if i%nCols == 0 {
			j = 0
			g = append(g, row{})
		}
		g[len(g)-1][j] = *it
		j++
	}

	// fmt.Println(len(g))
	// fmt.Println(g[0][0].Title)
	// fmt.Println(g[21][1].Title)

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, g)
	check(err)

	err = sendEmail("cjauvin@gmail.com", "cjauvin@gmail.com", "Netflix Fetcher", tpl.String(), *pw)
	check(err)
}
