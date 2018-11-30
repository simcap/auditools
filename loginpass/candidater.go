package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Candidater struct {
	post       *POST
	usernames  []string
	passwords  []string
	candidates []string
}

func (c *Candidater) Run() error {
	baseSig, err := c.Try("r32ftyior@f2hou27.com", "&Ui999qowfqfho")
	if err != nil {
		return err
	}

	for _, user := range c.usernames {
		for _, pass := range c.passwords {
			s, err := c.Try(user, pass)
			if err != nil {
				return err
			}
			if s.IsCandidate(baseSig) {
				c.candidates = append(c.candidates, fmt.Sprintf("%s|%s", user, pass))
			}
			wait := (10 + rand.Intn(6))
			time.Sleep(time.Duration(wait) * time.Second)
		}
	}

	return nil
}

func (c *Candidater) Try(username, pass string) (*Signature, error) {
	token, cookie, err := c.fetchAuthenticityTokenAndCookie()
	if err != nil {
		return nil, fmt.Errorf("grabing cookie and token: %s", err)
	}
	c.post.TokenVal = token

	u, err := url.ParseRequestURI(c.post.URL)
	if err != nil {
		return nil, err
	}

	u.Path = c.post.ActionPath
	verbose("Posting at %s", u)

	var body string
	if c.post.ContentType == "application/json" {
		data := make(map[string]interface{})
		data[c.post.Username] = username
		data[c.post.Password] = pass
		for _, input := range c.post.ExtraInputs {
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
			form.Set(c.post.TokenName, c.post.TokenVal)
		}
		form.Set(c.post.Username, username)
		form.Set(c.post.Password, pass)

		for _, input := range c.post.ExtraInputs {
			form.Set(input.Name, input.Value)
		}
		body = form.Encode()
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(body))
	req.Header.Add("Content-Length", strconv.Itoa(len(body)))
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:63.0) Gecko/20100101 Firefox/63.0")
	if c.post.ContentType == "application/json" {
		req.Header.Add("Content-Type", "application/json")
	} else {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if c.post.Referer != "" {
		req.Header.Add("Referer", c.post.Referer)
	}

	if cookie != nil {
		verbose("Set cookie %s=%s", cookie.Name, cookie.Value)
		req.AddCookie(cookie)
	}

	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	verbose("------------------------------------------------\n%s\n--------------------------------------------", dump)

	var serverStart, serverDone time.Time
	trace := &httptrace.ClientTrace{
		WroteRequest:         func(info httptrace.WroteRequestInfo) { serverStart = time.Now() },
		GotFirstResponseByte: func() { serverDone = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	var redirectCount int
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			redirectCount++
			verbose("Redirecting to %s (%d)", req.URL, len(via))
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if len(b) < 200 {
		verbose("----------- Response---------\n%s\n-----------------------------", b)
	} else {
		ioutil.WriteFile("response.html", b, 0666)
	}

	sig := &Signature{
		RedirectCount:        redirectCount,
		StatusCode:           resp.StatusCode,
		ResponseSize:         len(b),
		ServerProcessingTime: serverDone.Sub(serverStart),
	}

	log.Println(sig)

	return sig, nil
}

func (c *Candidater) fetchAuthenticityTokenAndCookie() (string, *http.Cookie, error) {
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
