package main

import "github.com/letsencrypt/gorepotemplate/config"

//import "errors"
import "flag"

//import "fmt"
import "log"

//import "os"

func main() {
	fs := flag.NewFlagSet("sunlight", flag.ExitOnError)
	configFlag := fs.String("config", "sunlight.yaml", "Path to YAML config file")

	c := config.Load_config_from_yaml(configFlag)

	logs := c.Logs
	seeds := []string{}

	for i := 0; i < len(logs); i++ {
		seeds = append(seeds, logs[i].Seed)
	}
	log.Println(seeds)

}
