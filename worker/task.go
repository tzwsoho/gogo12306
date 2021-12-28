package worker

import (
	"net/http/cookiejar"
	"time"
)

type TaskCB func(jar *cookiejar.Jar, task *Task) (err error)

type Task struct {
	QueryOnly bool

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

	Passengers  []string
	AllowPartly bool

	NextQueryTime time.Time
	CB            TaskCB
}
