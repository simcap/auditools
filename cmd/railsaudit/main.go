package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	railsDirFlag string
	dirs         = make(map[string]string)
)

type counter struct {
	inter [2]int
	val   int
}

type distribution struct {
	counters []*counter
	max      int
}

func newCounter(a, b int) *counter {
	return &counter{inter: [2]int{a, b}}
}

func (c *counter) isIn(i int) bool {
	if i >= c.inter[0] && i <= c.inter[1] {
		return true
	}
	return false
}

func (c *counter) String() string {
	max := c.inter[1]
	if max > 1e8 {
		return fmt.Sprintf("%d %6s", c.inter[0], "...")
	}
	return fmt.Sprintf("%5d - %4d", c.inter[0], max)
}

func newDistribution() *distribution {
	d := &distribution{
		counters: make([]*counter, 0),
	}
	for i := 0; i < 1000; i = i + 100 {
		d.counters = append(d.counters, newCounter(i, i+99))
	}
	for i := 1000; i < 10000; i = i + 1000 {
		d.counters = append(d.counters, newCounter(i, i+999))
	}
	d.counters = append(d.counters, newCounter(10000, 1e9))

	return d
}

func (d *distribution) print(title string, w io.Writer) {
	var lines []string
	lines = append(lines, "=============\n", fmt.Sprintf("= %s (%d lines)\n", title, d.sum()), "=============\n")
	for _, count := range d.counters {
		normal := (count.val * 160) / d.max
		lines = append(lines, fmt.Sprintf("%s | %s\n", count, strings.Repeat("*", normal)))
	}
	for _, l := range lines {
		fmt.Fprintf(w, l)
	}
	fmt.Fprintln(w)
}

func (d *distribution) sum() (sum int) {
	for _, count := range d.counters {
		sum = sum + count.val
	}
	return
}

func (d *distribution) put(i int) (sum int) {
	for _, count := range d.counters {
		if count.isIn(i) {
			count.val = count.val + i
			if count.val > d.max {
				d.max = count.val
			}
		}
	}
	return
}

func main() {
	flag.StringVar(&railsDirFlag, "d", ".", "Full path of the rails dir")
	flag.Parse()
	log.SetFlags(0)

	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	homeDir, err := filepath.Abs(railsDirFlag)
	check(err)

	for _, name := range []string{"app", "db", "config", "bin", "spec"} {
		d := filepath.Join(homeDir, name)
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			dirs[name] = d
		}
	}

	if len(dirs) != 5 {
		log.Fatalf("No Rails application at %s", homeDir)
	}

	type keepFunc func(info os.FileInfo) bool

	filter := func(dir string, f keepFunc) (out []string) {
		check(filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if f(info) {
				out = append(out, path)
			}
			return nil
		}))
		return out
	}

	rubies := filter(dirs["app"], func(info os.FileInfo) bool {
		return !info.IsDir() && filepath.Ext(info.Name()) == ".rb"
	})

	specs := filter(dirs["spec"], func(info os.FileInfo) bool {
		return !info.IsDir() && strings.HasSuffix(info.Name(), "_spec.rb")
	})

	log.Printf("%d ruby files, %d spec files", len(rubies), len(specs))
	log.Printf("%.2f files ratio", percent(len(specs), len(rubies)))

	rubiesDist := newDistribution()
	for _, s := range rubies {
		c, err := countLines(s)
		check(err)
		rubiesDist.put(c)
	}

	specsDist := newDistribution()
	for _, s := range specs {
		c, err := countLines(s)
		check(err)
		specsDist.put(c)
	}

	log.Printf("%d ruby lines, %d spec lines", rubiesDist.sum(), specsDist.sum())
	log.Printf("%.2f lines ratio", percent(specsDist.sum(), rubiesDist.sum()))

	rubiesDist.print("ruby", os.Stdout)
	specsDist.print("spec", os.Stdout)
}

func percent(a, b int) float64 {
	return float64(a*100) / float64(b)
}

func countLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	var sum int
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if l := bytes.TrimSpace(scanner.Bytes()); len(l) == 0 || l[0] == '#' {
			continue
		}
		sum++
	}
	return sum, scanner.Err()
}
