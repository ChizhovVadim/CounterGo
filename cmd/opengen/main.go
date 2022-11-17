package main

import (
	"context"
	"flag"
	"log"
)

type Config struct {
	outputPath string
	skip       int
	take       int
	ply        int
	seed       int
}

var config Config

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&config.outputPath, "output", "", "Path to output epd file")
	flag.IntVar(&config.skip, "skip", 0, "Skip")
	flag.IntVar(&config.take, "take", 300000, "Take")
	flag.IntVar(&config.ply, "ply", 8, "Ply")
	flag.IntVar(&config.seed, "seed", 0, "Seed")
	flag.Parse()

	log.Printf("%+v", config)

	var err = generateOpeningsRandomPipeline(context.Background(), config.outputPath, config.ply)
	if err != nil {
		log.Println(err)
	}
}
