package main

import (
	"fmt"
	"os"

	"github.com/firebolt-db/otel-exporter/cmd"
)

func main() {
	if err := cmd.NewApp().Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
