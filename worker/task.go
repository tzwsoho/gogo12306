package worker

import (
	"fmt"
	"net/http/cookiejar"
	"time"

	"go.uber.org/zap/zapcore"
)

type PassengerInfo struct {
	PassengerName string `json:"passenger_name"`         // 乘客姓名
	PassengerType int    `json:"passenger_type,string"`  // 乘客类型：1 - 成人票，2 - 儿童票，3 - 学生票，4 - 残军票
	SexName       string `json:"sex_name"`               // 性别
	BornDate      string `json:"born_date"`              // 出生日期
	IDTypeCode    string `json:"passenger_id_type_code"` // 证件类型代号：1 - 二代身份证，2 - 一代身份证，3 - 临时身份证，B - 护照，H - 外国人居留证，C - 港澳居民来往内地通行证，G - 台湾居民来往大陆通行证
	IDNumber      string `json:"passenger_id_no"`        // 证件号码
	MobileNumber  string `json:"mobile_no"`              // 手机号（有可能为空）
	UUID          string `json:"passenger_uuid"`         // 乘客全球唯一 ID，可以用来区别重名乘客
	AllEncStr     string `json:"allEncStr"`              // 乘客对应的密钥
}

func (p *PassengerInfo) String() string {
	return fmt.Sprintf("UUID: %s, 证件号码: %s, 姓名: %s", p.UUID, p.IDNumber, p.PassengerName)
}

type PassengerInfos []*PassengerInfo

func (p PassengerInfos) MarshalLogArray(arr zapcore.ArrayEncoder) (err error) {
	for _, passenger := range p {
		arr.AppendString(passenger.String())
	}

	return
}

type PassengerTicketInfo struct {
	PassengerInfo
	SeatType string // 坐席类型代号：
	BedPos   int    // 卧铺位置：0 - 不限，3 - 上铺，2 - 中铺，1 - 下铺
}

type TaskCB func(jar *cookiejar.Jar, task *Task) (err error)

type Task struct {
	QueryOnly bool

	OrderType      int
	OrderCandidate bool

	From string
	To   string

	FromTelegramCode string
	ToTelegramCode   string

	StartDates []string    // 出发日期
	SaleTimes  []time.Time // 开售时间

	TrainCodes []string

	Seats       []string
	SeatTypes   []int
	SeatIndices []int
	AllowNoSeat bool

	Passengers  PassengerInfos
	AllowPartly bool

	NextQueryTime time.Time
	CB            TaskCB
}
