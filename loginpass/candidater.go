package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

// Candidater uses a Poster to try username/pass combination.
// It the compares each try against a base negative response signature
// calculated at the start of a run.
//
// It then output potential candidates of valid credentials
type Candidater struct {
	poster           Poster
	waitTime, jitter int
	usernames        []string
	passwords        []string
	candidates       []string
	logger           *log.Logger
}

func NewCandidater(poster Poster) *Candidater {
	return &Candidater{
		poster:   poster,
		waitTime: 5,
		jitter:   5,
		logger:   log.New(os.Stdout, "", log.Ltime),
	}
}

func (c *Candidater) EstimatedMaxTime() int {
	return (len(c.usernames) * len(c.passwords) * (c.waitTime + c.jitter)) / 60
}

func (c *Candidater) Run() error {
	baseSig, err := c.poster.Try(randString(8), randString(13))
	if err != nil {
		return err
	}

	for _, user := range c.usernames {
		for _, pass := range c.passwords {
			s, err := c.poster.Try(user, pass)
			if err != nil {
				return err
			}
			c.logger.Printf("(%s|%s) %s", user, pass, s)

			if s.IsCandidate(baseSig) {
				c.candidates = append(c.candidates, fmt.Sprintf("%s|%s", user, pass))
			}
			time.Sleep(c.wait() * time.Second)
		}
	}

	return nil
}

func (c *Candidater) wait() time.Duration {
	wait := c.waitTime + rand.Intn(c.jitter)
	return time.Duration(wait)
}

var source = rand.NewSource(time.Now().UnixNano())

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[source.Int63()%int64(len(charset))]
	}
	return string(b)
}
