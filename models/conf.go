package models

import (
	"flag"
	"github.com/larspensjo/config"
	"log"
	"runtime"
)

func init(){
	myConfig := new(Config)
	config := myConfig.InitConfig("./", "EsConfig.ini", "ES")
	ES_HOST= config["es_host"]
	ES_PORT = config["es_port"]
	ES_INDEX_PREFIX = config["es_index_prefix"]
	ES_TYPE = config["es_type"]
	LOG_PATH=config["log_path"]

}

var (
		ES_HOST string
		ES_PORT string
		LOG_PATH string
		ES_INDEX_PREFIX string
		ES_TYPE string
		ES_INDEX string
		)
type Config struct {

}

func (c *Config) InitConfig(dir string,confName string,topic string) map[string]string {
	var (
		configFile = flag.String(dir, confName, "General configuration file")
	)
	//var topic string  = "nats"
	//topic list
	var TOPIC = make(map[string]string)
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	//set config file std
	cfg, err := config.ReadDefault(*configFile)
	if err != nil {
		log.Fatalf("Fail to find", *configFile, err)
	}
	//set config file std End

	//Initialized topic from the configuration
	if cfg.HasSection(topic) {
		section, err := cfg.SectionOptions(topic)
		if err == nil {
			for _, v := range section {
				options, err := cfg.String(topic, v)
				if err == nil {
					TOPIC[v] = options
				}
			}
		}
	}
	//Initialized topic from the configuration END

	return TOPIC
}