package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/simcap/auditools/passwords"
)

var (
	urlFlag          string
	basicAuthFlag    bool
	usernameListFlag string
	passwordListFlag string
	passwordDepth    int
	jitterFlag       int
	waitTimeFlag     int
	verboseFlag      bool
)

func main() {
	flag.StringVar(&urlFlag, "url", "", "URL of the resource")
	flag.BoolVar(&basicAuthFlag, "basicauth", false, "Basic authentication mode only (need url param)")
	flag.IntVar(&passwordDepth, "passdepth", 0, "Level of passwords generations & permutations")
	flag.StringVar(&usernameListFlag, "usernames", "admin", "Comma separated list of given usernames")
	flag.StringVar(&passwordListFlag, "passwords", "", "Comma separated list of given password (overwrite password generation)")
	flag.IntVar(&waitTimeFlag, "wait", 5, "Wait time in seconds between 2 requests")
	flag.IntVar(&jitterFlag, "jitter", 5, "Jitter interval in seconds to randomize wait time between requests")
	flag.BoolVar(&verboseFlag, "v", false, "Verbose mode")

	flag.Parse()
	log.SetFlags(0)

	var poster Poster
	if basicAuthFlag {
		if urlFlag == "" {
			log.Fatal(errors.New("missing url param when using basic auth"))
		}
		poster = &basicAuthPoster{urlFlag}
	} else {
		postForm, err := createPOSTForm(urlFlag)
		if err != nil {
			log.Fatal(err)
		}
		poster = &formPoster{postForm}
	}

	candidater := NewCandidater(poster)
	candidater.usernames = strings.Split(usernameListFlag, ",")
	if passwordListFlag != "" {
		candidater.passwords = strings.Split(passwordListFlag, ",")
	} else {
		options := passwords.Options{OrgOrURL: urlFlag}
		candidater.passwords = passwords.Generate(options)
	}
	candidater.waitTime = waitTimeFlag
	candidater.jitter = jitterFlag

	log.Printf("\nEstimated max time %d mins (wait time: %d, jitter: %d, usernames: %d, password count: %d)", candidater.EstimatedMaxTime(), waitTimeFlag, jitterFlag, len(candidater.usernames), len(candidater.passwords))

	if confirm() {
		if err := candidater.Run(); err != nil {
			log.Fatal(err)
		}
		log.Printf("Candidates: %v", candidater.candidates)
	}
}

func createPOSTForm(url string) (*POST, error) {
	if url == "" {
		return nil, errors.New("create form: missing url")
	}
	res, err := http.Get(url)
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
	doc.Find("input[type='password']").Each(func(i int, s *goquery.Selection) {
		s.Parents().Each(func(i int, s *goquery.Selection) {
			if goquery.NodeName(s) == "form" {
				post.ActionPath, _ = s.Attr("action")
				form = s
			}
		})
	})

	if form == nil {
		return nil, fmt.Errorf("no form found at %s", urlFlag)
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

	log.Println("created POST form")
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	enc.Encode(post)

	return post, nil
}

func confirm() bool {
	var s string
	fmt.Printf("\nRun (y/N)? ")
	if _, err := fmt.Scan(&s); err != nil {
		log.Fatal(err)
	}

	s = strings.TrimSpace(strings.ToLower(s))

	return s == "y" || s == "yes"
}

func verbose(msg string, a ...interface{}) {
	if verboseFlag {
		log.Printf(msg, a...)
	}
}
