package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"strings"
	"time"
)

type basicAuthPoster struct {
	url string
}

func (ba *basicAuthPoster) Try(username, pass string) (*Signature, error) {
	req, err := http.NewRequest("GET", ba.url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, pass)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:63.0) Gecko/20100101 Firefox/63.0")

	sig, resp, err := sendRequest(req)
	sig.Username = username

	canonical := http.CanonicalHeaderKey("WWW-Authenticate")
	if header := resp.Header.Get(canonical); resp.StatusCode != 200 && !strings.HasPrefix(header, "Basic") {
		return nil, fmt.Errorf("Not basic authentication as %s does not respond as basic auth (header %s=%q)", ba.url, canonical, header)
	}

	return sig, err
}

func sendRequest(req *http.Request) (*Signature, *http.Response, error) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, nil, err
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
		return nil, resp, err
	}

	verbose("-> Response %s\n", resp.Status)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}
	if len(b) < 200 {
		verbose("----------- Response Body ---------\n%s\n-----------------------------", b)
	}

	sig := &Signature{
		RedirectCount:        redirectCount,
		StatusCode:           resp.StatusCode,
		ResponseSize:         len(b),
		ServerProcessingTime: serverDone.Sub(serverStart),
	}

	return sig, resp, nil
}
