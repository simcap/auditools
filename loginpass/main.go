package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/simcap/sectools/passwords"
)

var (
	urlFlag          string
	formFileFlag     string
	basicAuthURLFlag string
	usernamesFlag    string
	jitterFlag       int
	waitTimeFlag     int
	verboseFlag      bool
)

func main() {
	flag.StringVar(&urlFlag, "url", "", "URL of the resource")
	flag.StringVar(&formFileFlag, "form-file", "", "Path of the form file to use")
	flag.StringVar(&basicAuthURLFlag, "basicauth-url", "", "URL of the basic auth")
	flag.StringVar(&usernamesFlag, "usernames", "", "Comma separated list of potential usernames")
	flag.IntVar(&waitTimeFlag, "wait", 5, "Wait time in seconds between 2 requests")
	flag.IntVar(&jitterFlag, "jitter", 5, "Jitter interval in seconds to randomize wait time between requests")
	flag.BoolVar(&verboseFlag, "v", false, "Verbose mode")

	flag.Parse()
	log.SetFlags(0)

	if urlFlag != "" {
		if err := createFormFile(); err != nil {
			log.Fatal(err)
		}
	}

	var poster Poster
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
		poster = &formPoster{post}
	}

	if basicAuthURLFlag != "" {
		poster = &basicAuthPoster{basicAuthURLFlag}
	}

	candidater := NewCandidater(poster)
	candidater.usernames = strings.Split(usernamesFlag, ",")
	candidater.passwords = passwords.Gen(poster.URL())
	candidater.waitTime = waitTimeFlag
	candidater.jitter = jitterFlag

	log.Printf("Estimated max time %d mins", candidater.EstimatedMaxTime())

	if err := candidater.Run(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Candidates: %v", candidater.candidates)
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
