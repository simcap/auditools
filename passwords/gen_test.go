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

	all, err := passwords.Gen(*stemFlag)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range all {
		fmt.Println(p)
	}
}
