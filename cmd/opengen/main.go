package main

import (
	"context"
	"flag"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var outputPath string
	flag.StringVar(&outputPath, "output", "", "Path to output epd file")
	flag.Parse()

	var err = generateOpeningsRandomPipeline(context.Background(), outputPath, 8)
	if err != nil {
		log.Println(err)
	}
}
