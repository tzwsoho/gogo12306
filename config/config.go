package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type ServerChan struct {
	On   bool   `json:"on"`
	SKey string `json:"skey"`
}

type Config struct {
	IsDevEnv       bool   `json:"is_dev_env"`
	LogFilepath    string `json:"log_filepath"`
	LogLevel       string `json:"log_level"`
	LogSplitMBSize int    `json:"log_split_mb_size"`
	LogKeepDays    int    `json:"log_keep_days"`

	CDNPath     string `json:"cdn_path"`
	GoodCDNPath string `json:"good_cdn_path"`

	OCRUrl string `json:"ocr_url"`

	ServerChan ServerChan `json:"serverchan,omitempty"`
}

var Cfg Config

func Init(cfgPath string) {
	if cfgData, err := ioutil.ReadFile(cfgPath); err != nil {
		log.Fatalf("read config err: %s", err.Error())
	} else {
		cfg := Config{}
		if err = json.Unmarshal(cfgData, &cfg); err != nil {
			log.Fatalf("config unmarshal err: %s", err.Error())
		} else {
			Cfg = cfg
		}
	}
}
