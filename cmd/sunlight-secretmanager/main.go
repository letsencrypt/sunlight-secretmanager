package main

import (
	"flag"
	"log"
	"os"

	"github.com/letsencrypt/sunlight-secretmanager/config"
)

func main() {
	fs := flag.NewFlagSet("sunlight", flag.ExitOnError)
	configFlag := fs.String("config", "foo", "Path to YAML config file")

	fs.Parse(os.Args[1:])

	c, err := config.LoadConfigFromYaml(configFlag)
	log.Println(c)
	log.Println(err)
}
