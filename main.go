package main

import (
	"k8s-event-listener/cmd"
	"os"
)

func main() {
	os.Exit(cmd.NewK8sEventListenerCommand().Run())
}
