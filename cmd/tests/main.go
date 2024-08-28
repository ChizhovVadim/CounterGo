package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ChizhovVadim/CounterGo/internal/utils"
)

func main() {
	var err = run(os.Args[1:])
	if err != nil {
		log.Println(err)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("command not specified")
	}
	var cmdName = args[0]
	args = args[1:]
	switch cmdName {
	//case "datasetsigmoid":
	//	return datasetSigmoidHandler(args)
	case "quality":
		return qualityHandler(args)
	case "tactic":
		return tacticHandler(args)
	case "arena":
		return arenaHandler(args)
	case "tuner":
		return tunerHandler(args)
	case "train":
		return trainHandler(args)
	case "perft":
		return perftHandler(args)
	case "play":
		return utils.PlayCli(newEngine(""))
	default:
		return fmt.Errorf("bad command %v", cmdName)
	}
}
