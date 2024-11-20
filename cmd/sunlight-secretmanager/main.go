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

	//actualSeeds := map[string]string{"2025h1b": "radiantlog-twig.ct.letsencrypt.org-2025h1b.key", "2025h2b": "radiantlog-twig.ct.letsencrypt.org-2025h2b.key"}
	returnedKeys, err := config.LoadAWSConfig(c)

	if err != nil {
		log.Printf("failed to load AWS config: [%v], err: [%v]", configFlag, err)
	} else {
		for key := range returnedKeys {
			log.Printf("successfully loaded key %v with value %v", key, returnedKeys[key])
		}
	}
}
