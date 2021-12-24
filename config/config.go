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
	GetCookieMethod int `json:"get_cookie_method"`

	ChromeDriverPath string `json:"chrome_driver_path"`

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

type WXPusher struct {
	On       bool     `json:"on"`
	AppToken string   `json:"app_token"`
	TopicIDs []int64  `json:"topic_ids"`
	UIDs     []string `json:"uids"`
}

type NotifierConfig struct {
	ServerChan `json:"serverchan,omitempty"`
	WXPusher   `json:"wxpusher,omitempty"`
}

type TaskConfig struct {
	From       string   `json:"from"`
	To         string   `json:"to"`
	Dates      []string `json:"dates"`
	Passengers []string `json:"passengers"`
	Seats      []string `json:"seats"`
	Trains     []string `json:"trains"`

	AllowInPart    bool `json:"allow_in_part"` // 允许部分提交
	AllowNoSeat    bool `json:"allow_no_seat"` // 允许提交系统分配的无座票
	QueryInterval  int  `json:"query_interval"`
	OrderCandidate bool `json:"order_candidate"` // 是否抢候补票
}

type Config struct {
	Logger   LoggerConfig   `json:"logger"`
	CDN      CDNConfig      `json:"cdn"`
	Login    LoginConfig    `json:"login"`
	OCR      OCRConfig      `json:"ocr"`
	Notifier NotifierConfig `json:"notifier"`
	Tasks    []TaskConfig   `json:"tasks"`
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
