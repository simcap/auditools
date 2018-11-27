package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	urlFlag      string
	formFileFlag string
	usernameFlag string
	passwordFlag string
	verboseFlag  bool
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

func main() {
	flag.StringVar(&urlFlag, "url", "", "URL of the resource")
	flag.StringVar(&formFileFlag, "form-file", "", "Path of the form file to use")
	flag.StringVar(&usernameFlag, "username", "", "Username to try")
	flag.StringVar(&passwordFlag, "pass", "azerty%%654", "Password to try")
	flag.BoolVar(&verboseFlag, "v", false, "Verbose mode")

	flag.Parse()
	log.SetFlags(0)

	if urlFlag != "" {
		if err := createFormFile(); err != nil {
			log.Fatal(err)
		}
	}

	if formFileFlag != "" {
		if err := tryLoginPass(); err != nil {
			log.Fatal(err)
		}
	}
}

func tryLoginPass() error {
	file, err := os.Open(formFileFlag)
	var post POST
	dec := json.NewDecoder(file)
	dec.Decode(&post)

	token, cookie, err := grabAuthenticityTokenAndCookie(post.URL)
	if err != nil {
		return fmt.Errorf("grabing cookie and token: %s", err)
	}
	post.TokenVal = token

	u, err := url.ParseRequestURI(post.URL)
	if err != nil {
		return err
	}

	u.Path = post.ActionPath
	verbose("Posting at %s", u)

	var body string
	if post.ContentType == "application/json" {
		data := make(map[string]interface{})
		data[post.Username] = usernameFlag
		data[post.Password] = passwordFlag
		for _, input := range post.ExtraInputs {
			data[input.Name] = input.Value
		}
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		body = string(b)
	} else {
		form := url.Values{}
		if token != "" {
			verbose("Set authenticity token %s", token)
			form.Set(post.TokenName, post.TokenVal)
		}
		form.Set(post.Username, usernameFlag)
		form.Set(post.Password, passwordFlag)

		for _, input := range post.ExtraInputs {
			form.Set(input.Name, input.Value)
		}
		body = form.Encode()
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(body))
	req.Header.Add("Content-Length", strconv.Itoa(len(body)))
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:63.0) Gecko/20100101 Firefox/63.0")
	if post.ContentType == "application/json" {
		req.Header.Add("Content-Type", "application/json")
	} else {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	if post.Referer != "" {
		req.Header.Add("Referer", post.Referer)
	}

	if cookie != nil {
		verbose("Set cookie %s=%s", cookie.Name, cookie.Value)
		req.AddCookie(cookie)
	}

	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return err
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
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(b) < 200 {
		verbose("----------- Response---------\n%s\n-----------------------------", b)
	} else {
		ioutil.WriteFile("response.html", b, 0666)
	}
	log.Printf("Redirect: %d, Status: %d, Length: %d, Server processing: %s\n", redirectCount, resp.StatusCode, len(b), serverDone.Sub(serverStart))

	return nil
}

func grabAuthenticityTokenAndCookie(url string) (string, *http.Cookie, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
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

func createFormFile() error {
	res, err := http.Get(urlFlag)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	csrfParam := doc.Find("head > meta[name='csrf-param']").First()
	csrfToken := doc.Find("head > meta[name='csrf-token']").First()

	param, _ := csrfParam.Attr("content")
	tok, _ := csrfToken.Attr("content")

	post := &POST{
		URL:       urlFlag,
		TokenName: param,
		TokenVal:  tok,
	}

	var form *goquery.Selection
	doc.Find(fmt.Sprintf("form input[name='%s']", param)).Each(func(i int, s *goquery.Selection) {
		current := s.Parents().First()
		action, _ := current.Attr("action")
		if strings.Contains(action, "sign") || strings.Contains(action, "login") {
			form = current
			post.ActionPath = action
		}
	})

	if form == nil {
		return fmt.Errorf("no form found at %s", urlFlag)
	}

	var inputs []Input
	form.Find("input").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		value, _ := s.Attr("value")
		inputs = append(inputs, Input{Name: name, Value: value})
	})

	for _, in := range inputs {
		if strings.Contains(in.Name, "pass") {
			post.Password = in.Name
			continue
		}
		if strings.Contains(in.Name, "log") || strings.Contains(in.Name, "name") || strings.Contains(in.Name, "mail") {
			post.Username = in.Name
			continue
		}
		if in.Value != "" && in.Name != "authenticity_token" {
			post.ExtraInputs = append(post.ExtraInputs, in)
		}

	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(post)

	return nil
}

func verbose(msg string, a ...interface{}) {
	if verboseFlag {
		log.Printf(msg, a...)
	}
}
