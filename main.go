package main

import (
	"context"
	"k8s-event-listener/cmd"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	rootApp := cmd.NewK8sEventListenerCommand(ctx)

	go func() {
		<-sigs
		cancel()
		os.Exit(1)
	}()

	res := rootApp.Run()
	cancel()
	os.Exit(res)
}
