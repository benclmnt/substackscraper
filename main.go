package main

import (
	"os"

	scraper "github.com/benclmnt/substackscraper/scraper"
)

func main() {
	os.Exit(scraper.CLI(os.Args[1:]))
}
