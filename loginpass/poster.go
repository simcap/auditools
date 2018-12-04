package main

import (
	"fmt"
	"time"
)

type Poster interface {
	Try(username, pass string) (*Signature, error)
}

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
