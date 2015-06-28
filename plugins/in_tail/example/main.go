package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/najeira/gigo/plugins/in_tail"
)

type emitter struct {
}

func (e *emitter) Emit(msg interface{}) error {
	if b, ok := msg.([]byte); ok {
		fmt.Println(string(b))
	} else if s, ok := msg.(string); ok {
		fmt.Println(s)
	} else {
		return fmt.Errorf("unknown type")
	}
	return nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	file := os.Args[1]
	e := emitter{}

	input := in_tail.New(in_tail.Config{
		File:    file,
		Emitter: &e,
		Logger:  nil,
	})

	if err := input.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	if err := input.Stop(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
