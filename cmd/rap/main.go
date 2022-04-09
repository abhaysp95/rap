package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	p "rap/parse"
)

// TODO:
// 1. Add option to print the result only for provided rollno.
// 2. Add option to print the semester only result for the provided rollno.
//    (back sem should be given a seperate flag)
// 3. Add option to store the result in json file
// 4. Try to use "cobra" for handling cli args
// 5. Put out basic "viper" configuration for handling configuration of "rap"

var dirHTML = flag.String("dir", ".", "provide directory with HTML files")

func main() {
	flag.Parse()
	files, err := filepath.Glob(filepath.Join(*dirHTML, "*.html"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(files)

	for _, htmlFile := range files {
		rf, err := os.Open(htmlFile)
		if err != nil {
			log.Fatal(err)
		}
		p.Parse(rf)
	}


	fmt.Println("Completed")
}
