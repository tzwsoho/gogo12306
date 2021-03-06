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

	ChromeBrowserPath string `json:"chrome_browser_path"`
	ChromeDriverPath  string `json:"chrome_driver_path"`

	RailExpiration string `json:"rail_expiration"`
	RailDeviceID   string `json:"rail_device_id"`

	Username string `json:"username"`
	Password string `json:"password"`

	LoginMethod int `json:"login_method"`

	OCRUrl string `json:"ocr_url"`

	CastNum string `json:"cast_num"`
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
	QueryOnly bool `json:"query_only"`

	OrderType int `json:"order_type"` // 1 - 普通购票，2 - 候补票/刷票
	BlackTime int `json:"black_time"`

	AllowCandidate    bool `json:"allow_candidate"`    // 是否抢候补票
	CandidateDeadline int  `json:"candidate_deadline"` // 候补票距离开车前的截止兑换时间

	From string `json:"from"`
	To   string `json:"to"`

	StartDates []string `json:"start_dates"`

	TrainCodes []string `json:"train_codes"`

	Seats          []string `json:"seats"`
	ChooseSeats    []string `json:"choose_seats"`
	SeatDetailType []string `json:"seat_detail_type"`
	AllowNoSeat    bool     `json:"allow_no_seat"` // 允许提交系统分配的无座票

	Passengers  []string `json:"passengers"`
	UUIDs       []string `json:"uuids"`
	AllowPartly bool     `json:"allow_partly"` // 允许部分提交
}

type Config struct {
	Logger   LoggerConfig   `json:"logger"`
	CDN      CDNConfig      `json:"cdn"`
	Login    LoginConfig    `json:"login"`
	Notifier NotifierConfig `json:"notifier"`
	Tasks    []TaskConfig   `json:"tasks"`

	StudentPresellDays int // 学生票预售提前天数
	OtherPresellDays   int // 一般车票预售提前天数
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

			// 暂定 15 天，后面的 NeedCaptcha 接口可以获取正确数值
			Cfg.StudentPresellDays = 15
			Cfg.OtherPresellDays = 15
		}
	}
}
