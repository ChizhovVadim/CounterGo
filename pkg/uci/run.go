package uci

import (
	"log"
	"os"
)

func Run(eng *EngineAgent) error {
	var cmds = make(chan any)
	defer close(cmds)
	go func() {
		var err = eng.Run(cmds)
		if err != nil {
			log.Println(err)
		}
	}()
	return ReadCommands(os.Stdin, cmds)
}
