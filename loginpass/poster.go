package main

import (
	"fmt"
	"time"
)

type Poster interface {
	Try(username, pass string) (*Signature, error)
}

type Signature struct {
	Username             string
	RedirectCount        int
	StatusCode           int
	ResponseSize         int
	ServerProcessingTime time.Duration
}

func (s Signature) String() string {
	return fmt.Sprintf("Redirects: %d, Status: %d, Length: %d, NormalizedLength: %d, ServerProcessing: %s", s.RedirectCount, s.StatusCode, s.ResponseSize, s.ResponseSize-2*len(s.Username), s.ServerProcessingTime)
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
