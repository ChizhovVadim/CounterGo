package uci

import (
	"bufio"
	"context"
	"log"
	"os"
)

type CommandHandler interface {
	Handle(ctx context.Context, command string) error
}

func RunCli(logger *log.Logger, handler CommandHandler) {
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	var scanner = bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var commandLine = scanner.Text()
		if commandLine == "quit" {
			return
		}
		var err = handler.Handle(ctx, commandLine)
		if err != nil {
			logger.Println(err)
		}
	}
}
