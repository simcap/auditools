package passwords_test

import (
	"flag"
	"fmt"
	"testing"

	"github.com/simcap/sectools/passwords"
)

var stemFlag = flag.String("stem", "https://facebook.com", "Stem used to generate some passwords")

func TestGenerateToExploreOutput(t *testing.T) {
	flag.Parse()

	fmt.Printf("Generating with stem %s\n\n", *stemFlag)

	for _, p := range passwords.Gen(*stemFlag) {
		fmt.Println(p)
	}
}
