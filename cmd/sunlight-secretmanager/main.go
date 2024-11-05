package main

import (
	"flag"
	"log"
)

import "github.com/letsencrypt/gorepotemplate/config"

func main() {
	fs := flag.NewFlagSet("sunlight", flag.ExitOnError)
	configFlag := fs.String("config", "sunlight.yaml", "Path to YAML config file")

	c := config.LoadConfigFromYaml(configFlag)

	logs := c.Logs
	seeds := []string{}

	for i := range logs {
		seeds = append(seeds, logs[i].Seed)
	}
	log.Println(seeds)
}
