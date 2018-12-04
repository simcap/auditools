package passwords_test

import (
	"flag"
	"fmt"
	"testing"

	"github.com/simcap/sectools/passwords"
)

var opts passwords.Options

func init() {
	flag.StringVar(&opts.OrgOrURL, "org-or-url", "", "Org or url to generate passwords from")
	flag.StringVar(&opts.Firstname, "firstname", "", "Firstname to generate passwords from")
}

func TestGenerateToExploreOutput(t *testing.T) {
	flag.Parse()

	for _, p := range passwords.Generate(opts) {
		fmt.Println(p)
	}
}
