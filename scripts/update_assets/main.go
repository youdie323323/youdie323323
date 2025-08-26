package main

import (
	"flag"

	"update_assets/language"
)

func main() {
	flag.Parse()

	language.ConstructLanguageInformationSVG()
}
