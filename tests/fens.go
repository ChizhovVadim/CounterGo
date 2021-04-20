package main

import (
	"bufio"
	"errors"
	"os"
	"time"
)

type testFenFunc func(fen string) error

func testFens(name string, f testFenFunc) {
	const MaxErrNb = 10
	var start = time.Now()
	var errNb = 0
	logger.Println("Test started", name)
	defer func() {
		var testPassed = errNb == 0
		if r := recover(); r != nil {
			testPassed = false
			logger.Println(r)
		}
		var elapsed = time.Since(start)
		if testPassed {
			logger.Println("Test finished", name, elapsed)
		} else {
			logger.Println("Test FAILED", name, elapsed)
		}
	}()
	file, err := os.Open(TuneFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var scanner = bufio.NewScanner(file)
	var index = 0
	for scanner.Scan() {
		index++
		var line = scanner.Text()
		var entry, err = parseLearnEntry(line)
		if err != nil {
			panic(err)
		}
		err = f(entry.fen)
		if err != nil {
			errNb++
			if errNb >= MaxErrNb {
				panic(errors.New("Max err number"))
			}
			logger.Printf("%v %v %v", index, entry.fen, err)
		}
	}
}
