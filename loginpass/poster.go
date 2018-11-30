package main

import "time"

type Signature struct {
	RedirectCount        int
	StatusCode           int
	ResponseSize         int
	ServerProcessingTime time.Duration
}

type Poster interface {
	Try(username, pass string) (*Signature, error)
	URL() string
}
