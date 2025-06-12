package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/symonk/vessel/cmd"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Println("error", err)
		os.Exit(1)
	}

}
