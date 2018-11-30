package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	urlFlag               string
	formFileFlag          string
	passwordsFilepathFlag string
	usernamesFilepathFlag string
	verboseFlag           bool
)

type Signature struct {
	RedirectCount        int
	StatusCode           int
	ResponseSize         int
	ServerProcessingTime time.Duration
}

func (s Signature) String() string {
	return fmt.Sprintf("Redirects: %d, Status: %d, Length: %d, ServerProcessing: %s", s.RedirectCount, s.StatusCode, s.ResponseSize, s.ServerProcessingTime)
}

func (s Signature) IsCandidate(base *Signature) bool {
	if s.RedirectCount != base.RedirectCount || s.StatusCode != base.StatusCode {
		return true
	}
	if (s.ResponseSize/base.ResponseSize)*10 > 11 {
		return true
	}
	return false
}

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
	flag.StringVar(&usernamesFilepathFlag, "usernames", "usernames", "Path to file containing multiline potential usernames")
	flag.StringVar(&passwordsFilepathFlag, "passwords", "passwords", "Path to file containing multiline potential passwords")
	flag.BoolVar(&verboseFlag, "v", false, "Verbose mode")

	flag.Parse()
	log.SetFlags(0)

	if urlFlag != "" {
		if err := createFormFile(); err != nil {
			log.Fatal(err)
		}
	}

	if formFileFlag != "" {
		file, err := os.Open(formFileFlag)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		post := new(POST)
		if err := json.NewDecoder(file).Decode(post); err != nil {
			log.Fatal(err)
		}

		candidater := &Candidater{
			post:      post,
			usernames: fileToArr(usernamesFilepathFlag),
			passwords: fileToArr(passwordsFilepathFlag),
		}

		log.Printf("Estimated max time %d mins", (len(candidater.usernames)*len(candidater.usernames)*15)/60)

		if err := candidater.Run(); err != nil {
			log.Fatal(err)
		}

		log.Printf("Candidates: %v", candidater.candidates)
	}
}

func fileToArr(path string) (out []string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		out = append(out, strings.TrimSpace(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
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
