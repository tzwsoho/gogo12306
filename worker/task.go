package worker

import (
	"gogo12306/common"
	"net/http/cookiejar"
	"time"
)

type TaskCB func(jar *cookiejar.Jar, task *Task) (err error)

type Task struct {
	TaskID    int64
	QueryOnly bool
	Done      chan struct{}

	OrderType int
	BlackTime int

	AllowCandidate    bool
	CandidateDeadline int

	From string
	To   string

	FromTelegramCode string
	ToTelegramCode   string

	StartDates []string    // 出发日期
	SaleTimes  []time.Time // 开售时间

	TrainCodes []string

	Seats          []string
	SeatTypes      []int
	SeatIndices    []int
	ChooseSeats    []string
	SeatDetailType []string
	AllowNoSeat    bool

	Passengers  common.PassengerInfos
	AllowPartly bool

	NextQueryTime time.Time
	CB            TaskCB
}
