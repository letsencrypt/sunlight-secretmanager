package main

import (
	"flag"
	"log"
	"os"

	"github.com/letsencrypt/sunlight-secretmanager/config"
)

func main() {
	fs := flag.NewFlagSet("sunlight", flag.ExitOnError)
	configFlag := fs.String("config", "", "Path to YAML config file")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Println("Error parsing flags", err)
		os.Exit(1)
	}

	c, err := config.LoadConfigFromYaml(*configFlag)
	if err != nil {
		log.Printf("failed to read or parse config file: [%v], err: [%v]", configFlag, err)
	} else {
		log.Printf("seeds: %v", c)
	}
}
