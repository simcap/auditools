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

type generator func(s string) []string

type transform func(g generator, s string) generator

func Gen(stem string) []string {
	var all []string

	funcs := []generator{
		capitalizer(
			genCommon,
			genKeyboardwalk,
			lightLeetSpeaker(genFromStem),
		),
	}

	unique := make(map[string]struct{})
	for _, g := range funcs {
		for _, pass := range g(stem) {
			unique[pass] = struct{}{}
		}
	}
	for p := range unique {
		all = append(all, p)
	}

	sort.Strings(all)

	return all
}

func capitalizer(gens ...generator) generator {
	return func(s string) (all []string) {
		for _, l := range concat(gens...)(s) {
			all = append(all, l)
			all = append(all, capitalize(l))
		}
		return
	}
}

func concat(gens ...generator) generator {
	return func(s string) (all []string) {
		for _, g := range gens {
			for _, l := range g(s) {
				all = append(all, l)
			}
		}
		return
	}
}

func lightLeetSpeaker(g generator) generator {
	replacer := strings.NewReplacer("o", "0", "i", "1", "e", "3")

	return func(s string) (all []string) {
		for _, l := range g(s) {
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

func genFromStem(stem string) (list []string) {
	if ip := net.ParseIP(stem); ip != nil {
		return
	}

	word := stem

	if u, _ := url.ParseRequestURI(stem); u != nil && u.Hostname() != "" {
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

func genCommon(string) []string {
	return []string{
		"12345678",
		"123456789",
		"qwertyuiop",
		"qwertyui",
		"asdfghjk",
		"passw0rd",
		"password123!",
		"password1234",
		"1qa2ws3ed4rf",
		"1q2w3e4r5t6y",
	}
}

func genKeyboardwalk(string) []string {
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
