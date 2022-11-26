package main

import (
	"log"
	"os"

	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
)

const configFile = "config/dev.yaml"

var c *conf

func init() {
	yamlFile, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalln("error in read configFIle:", err.Error())
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalln("error in unmarshal configFIle:", err.Error())
	}
}

func main() {
	b, err := NewBot(c)
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = b.StartBot()
	if err != nil {
		log.Fatalln(err.Error())
	}
}
