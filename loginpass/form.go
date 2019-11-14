package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Input struct {
	Name, Value string
}

type POST struct {
	URL         string
	Referer     string
	ActionPath  string
	ContentType string
	Username    string
	Password    string
	TokenName   string
	TokenVal    string
	ExtraInputs []Input
}

type formPoster struct {
	post *POST
}

func (fp *formPoster) Try(username, pass string) (*Signature, error) {
	token, cookie, err := fp.refreshAuthenticityTokenAndCookie()
	if err != nil {
		return nil, fmt.Errorf("grabing cookie and token: %s", err)
	}
	fp.post.TokenVal = token

	u, err := url.ParseRequestURI(fp.post.URL)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(fp.post.ActionPath, "http") {
		u.Path = fp.post.ActionPath
	}
	verbose("Posting at %s", u)

	var body string
	if fp.post.ContentType == "application/json" {
		data := make(map[string]interface{})
		data[fp.post.Username] = username
		data[fp.post.Password] = pass
		for _, input := range fp.post.ExtraInputs {
			data[input.Name] = input.Value
		}
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = string(b)
	} else {
		form := url.Values{}
		if token != "" {
			verbose("Set authenticity token %s", token)
			form.Set(fp.post.TokenName, fp.post.TokenVal)
		}
		form.Set(fp.post.Username, username)
		form.Set(fp.post.Password, pass)

		for _, input := range fp.post.ExtraInputs {
			form.Set(input.Name, input.Value)
		}
		body = form.Encode()
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(body))
	req.Header.Add("Content-Length", strconv.Itoa(len(body)))
	req.Header.Add("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:63.0) Gecko/20100101 Firefox/63.0")
	if fp.post.ContentType == "application/json" {
		req.Header.Add("Content-Type", "application/json")
	} else {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if fp.post.Referer != "" {
		req.Header.Add("Referer", fp.post.Referer)
	}

	if cookie != nil {
		verbose("Set cookie %s=%s", cookie.Name, cookie.Value)
		req.AddCookie(cookie)
	}

	sig, _, err := sendRequest(req)
	sig.Username = username

	return sig, err
}

func (c *formPoster) refreshAuthenticityTokenAndCookie() (string, *http.Cookie, error) {
	res, err := http.Get(c.post.URL)
	if err != nil {
		return "", nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", nil, nil
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", nil, err
	}

	csrfToken := doc.Find("head > meta[name='csrf-token']").First()
	tok, _ := csrfToken.Attr("content")

	var cookie *http.Cookie
	for _, c := range res.Cookies() {
		cookie = c
		break
	}

	return tok, cookie, nil
}
