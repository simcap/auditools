package passwords

import (
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"
	"time"
)

type generator func(s string) []string

type transform func(g generator, s string) generator

func Gen(stem string) []string {
	var all []string

	funcs := []generator{
		genUSCommon,
		genUSKeyboardwalk,
		tranToDigits(genFromURL),
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

func tranToDigits(g generator) generator {
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

func genFromURL(websiteURL string) (list []string) {
	u, err := url.ParseRequestURI(websiteURL)
	if err != nil {
		return
	}
	host := strings.ToLower(u.Hostname())

	if ip := net.ParseIP(host); ip != nil {
		return
	}

	if last := strings.LastIndex(host, "."); last > 0 {
		host = host[0:last]
		if last := strings.LastIndex(host, "."); last > 0 && last < len(host) {
			host = host[last+1:]
		}
	}

	today := time.Now()
	for _, s := range []string{host, capitalize(host)} {
		list = append(list, fmt.Sprintf("%s%d", s, today.Year()))
		list = append(list, fmt.Sprintf("%s%d", s, today.AddDate(-1, 0, 0).Year()))
		list = append(list, fmt.Sprintf("%s%d", s, today.AddDate(-2, 0, 0).Year()))
	}

	return
}

func genUSCommon(string) []string {
	return []string{
		"12345678",
		"123456789",
		"qwertyuiop",
		"qwertyui",
		"asdfghjk",
		"passw0rd",
		"Password123!",
		"1qa2ws3ed4rf",
		"1q2w3e4r5t6y",
	}
}

func genUSKeyboardwalk(string) []string {
	return []string{
		"1qa2ws3ed4rf",
		"1q2w3e4r5t6y",
	}
}

func capitalize(s string) string {
	if len(s) > 0 {
		return fmt.Sprintf("%s%s", strings.ToUpper(string(s[0])), s[1:])
	}
	return s
}
