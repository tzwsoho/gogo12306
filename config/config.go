package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type LoggerConfig struct {
	IsDevelop      bool   `json:"is_develop"`
	LogLevel       string `json:"log_level"`
	LogFilepath    string `json:"log_filepath"`
	LogSplitMBSize int    `json:"log_split_mb_size"`
	LogKeepDays    int    `json:"log_keep_days"`
}

type CDNConfig struct {
	CDNPath     string `json:"cdn_path"`
	GoodCDNPath string `json:"good_cdn_path"`
}

type LoginConfig struct {
	GetWebBrowserIDType int `json:"get_webbroswer_id_type"`

	PjialinUrl string `json:"pjialin_url"`

	RailExpiration string `json:"rail_expiration"`
	RailDeviceID   string `json:"rail_device_id"`

	Username string `json:"username"`
	Password string `json:"password"`
}

type OCRConfig struct {
	OCRUrl string `json:"ocr_url"`
}

type ServerChan struct {
	On   bool   `json:"on"`
	SKey string `json:"skey"`
}

type NotifierConfig struct {
	ServerChan `json:"serverchan,omitempty"`
}

type Config struct {
	Logger   LoggerConfig   `json:"logger"`
	CDN      CDNConfig      `json:"cdn"`
	Login    LoginConfig    `json:"login"`
	OCR      OCRConfig      `json:"ocr"`
	Notifier NotifierConfig `json:"notifier"`
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
