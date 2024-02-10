package main

import (
	"log"
	"os"
)

var cliArgs = NewCommandArgs(os.Args)

func main() {
	var err = run()
	if err != nil {
		log.Println(err)
	}
}

func run() error {
	var cli = NewCommandHandler()
	cli.Add("tuner", tunerHandler)
	cli.Add("train", trainHandler)
	cli.Add("quality", qualityHandler)
	return cli.Execute(cliArgs.CommandName())
}
