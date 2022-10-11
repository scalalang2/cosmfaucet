package main

import (
	"flag"
	"log"
	"os"

	"github.com/scalalang2/cosmfaucet/core"
)

var flagConfig = flag.String("config-file", "config.yaml", "path to config file")

func init() {
	flag.Parse()
	if *flagConfig == "" {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	rootConfig := core.LoadConfig(*flagConfig)
	app, err := core.NewApp(rootConfig)
	if err != nil {
		log.Fatalf("failed to initialize the application: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("failed to run the application: %v", err)
	}

	<-make(chan bool)
}
