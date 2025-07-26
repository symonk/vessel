package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"net"

	"github.com/symonk/vessel/cmd"
)

// TODO: Remove me later
func init() {
	net.DefaultResolver.PreferGo = true
}

func main() {
	//defer profiler.Start(profiler.WithHeapProfiler()).Stop()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()
	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
