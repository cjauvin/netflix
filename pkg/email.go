package lib

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"
)

const (
	nCols = 5
)

type row [nCols]Item

func SendEmail(from string, to string, subject string, body string, pw string) (err error) {

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	fr := fmt.Sprintf("From: %s\n", from)
	t := fmt.Sprintf("To: %s\n", to)
	sub := fmt.Sprintf("Subject: %s\n", subject)
	msg := []byte(fr + t + sub + mime + body)

	auth := smtp.PlainAuth("", from, pw, "smtp.gmail.com")
	err = smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, msg)
	return
}

func BuildEmailBody(items []*Item) string {

	tmpl := template.Must(template.New("item_table").Parse(`
<table>
  {{range .}}
    <tr>
      {{range .}}
      <td>
        {{if .ItemID }}
	  <img src={{.ImageUrl}}><img>
	  <br />
	  <i>{{.Title}}</i>
	  <hr />
	  {{.Summary}}
          <br />
	  <a href="http://imdb.com/title/{{.ImdbID}}">IMDb</a>
	  <br />
	  <a href="https://netflix.com/title/{{.NetflixID}}">Netflix</a>
        {{end}}
      </td>
      {{end}}
    </tr>
  {{end}}
</table>`))

	g := []row{}

	var j int
	for i, it := range items {
		if i%nCols == 0 {
			j = 0
			g = append(g, row{})
		}
		g[len(g)-1][j] = *it
		j++
	}

	var tpl bytes.Buffer
	err := tmpl.Execute(&tpl, g)
	Check(err)

	return tpl.String()
}
