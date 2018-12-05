package passwords

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"
)

type generatorFunc func(g Options) []string

type transformFunc func(g generatorFunc, s string) generatorFunc

type Options struct {
	Firstname string
	OrgOrURL  string
}

func Generate(opts Options) []string {
	var all []string

	funcs := []generatorFunc{
		capitalizer(genCommon, genKeyboardwalk),
	}

	if opts.Firstname != "" {
		funcs = append(funcs,
			capitalizer(lightLeetSpeaker(genFromFirstname)),
		)
	}

	if opts.OrgOrURL != "" {
		funcs = append(funcs,
			capitalizer(lightLeetSpeaker(genFromOrgOrURL)),
		)
	}

	unique := make(map[string]struct{})
	for _, g := range funcs {
		for _, pass := range g(opts) {
			unique[pass] = struct{}{}
		}
	}
	for p := range unique {
		all = append(all, p)
	}

	sort.Strings(all)

	return all
}

func capitalizer(gens ...generatorFunc) generatorFunc {
	return func(opts Options) (all []string) {
		for _, l := range concat(gens...)(opts) {
			all = append(all, l)
			all = append(all, capitalize(l))
		}
		return
	}
}

func concat(gens ...generatorFunc) generatorFunc {
	return func(opts Options) (all []string) {
		for _, g := range gens {
			for _, l := range g(opts) {
				all = append(all, l)
			}
		}
		return
	}
}

func lightLeetSpeaker(g generatorFunc) generatorFunc {
	replacer := strings.NewReplacer("o", "0", "i", "1", "e", "3")

	return func(opts Options) (all []string) {
		for _, l := range g(opts) {
			all = append(all, l)
			all = append(all,
				strings.Replace(l, "o", "0", -1),
			)
			all = append(all,
				strings.Replace(l, "i", "1", -1),
			)
			all = append(all,
				strings.Replace(l, "e", "3", -1),
			)

			all = append(all, replacer.Replace(l))
		}

		return
	}
}

const (
	maxAge = 50
	minAge = 30
)

func genFromFirstname(opts Options) (list []string) {
	name := opts.Firstname
	year := time.Now().Year()
	for i := year - maxAge; i <= year-minAge; i++ {
		list = append(list, fmt.Sprintf("%s%d", name, i))
	}
	return
}

func genFromOrgOrURL(opts Options) (list []string) {
	if ip := net.ParseIP(opts.OrgOrURL); ip != nil {
		return
	}

	word := opts.OrgOrURL

	if u, _ := url.ParseRequestURI(opts.OrgOrURL); u != nil && u.Hostname() != "" {
		host := strings.ToLower(u.Hostname())
		if last := strings.LastIndex(host, "."); last > 0 {
			host = host[0:last]
			if last := strings.LastIndex(host, "."); last > 0 && last < len(host) {
				host = host[last+1:]
			}
		}
		word = host
	}

	today := time.Now()
	list = append(list, fmt.Sprintf("%s%d", word, today.Year()))
	list = append(list, fmt.Sprintf("%s%d", word, today.AddDate(-1, 0, 0).Year()))
	list = append(list, fmt.Sprintf("%s%d", word, today.AddDate(-2, 0, 0).Year()))

	return
}

func genCommon(Options) []string {
	return []string{
		"12345678",
		"123456789",
		"qwertyuiop",
		"qwertyui",
		"asdfghjk",
		"password",
		"password123456",
		"password123456789",
		"password987654321",
		"password1234",
		"password123!",
		"1qa2ws3ed4rf",
		"1q2w3e4r5t6y",
	}
}

func genKeyboardwalk(Options) []string {
	return []string{
		// qwerty keyboard
		"1qa2ws3ed4rf",
		"1q2w3e4r5t6y",
		// azerty keyboard
		"1aq2sz3de4rf",
		"1a2z3e4r5t6y",
	}
}

func capitalize(s string) string {
	if len(s) > 0 {
		if first := s[0]; unicode.IsLetter(rune(first)) {
			return fmt.Sprintf("%s%s", strings.ToUpper(string(first)), s[1:])
		}
	}
	return s
}
