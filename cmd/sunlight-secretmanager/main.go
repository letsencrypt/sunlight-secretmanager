package main

import (
	"flag"
	"log"

	"github.com/letsencrypt/sunlight-secretmanager/config"
)

func main() {
	fs := flag.NewFlagSet("sunlight", flag.ExitOnError)
	configFlag := fs.String("config", "sunlight.yaml", "Path to YAML config file")

	c, err := config.LoadConfigFromYaml(configFlag)
	log.Println(c)
	log.Println(err)
}
