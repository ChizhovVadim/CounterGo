package main

import (
	"context"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var err = generateOpenings()
	if err != nil {
		log.Println(err)
	}
}

func generateOpenings() error {
	/*return generateOpeningsPipeline(context.Background(),
	4,
	"/Users/vadimchizhov/chess/millionbase-2.5.pgn",
	"/Users/vadimchizhov/chess/openings.epd",
	14)*/
	return generateOpeningsRandomPipeline(context.Background(),
		"/Users/vadimchizhov/chess/openings_random.epd",
		8)
}
