package worker

import (
	"errors"
	"gogo12306/config"
	"gogo12306/station"
	"sort"
	"time"
)

type Task struct {
	NextQueryTime time.Time

	From string
	To   string

	SaleTime     []time.Time
	TrainNumbers []string
	Seats        []string
	Passengers   []string
}

func (t *Task) Parse(task *config.TaskConfig) (err error) {
	from := station.StationNameToStationInfo(task.From)
	if from == nil {
		return errors.New("from error")
	} else {
		t.From = from.TelegramCode
	}

	to := station.StationNameToStationInfo(task.To)
	if to == nil {
		return errors.New("to error")
	} else {
		t.To = to.TelegramCode
	}

	if len(task.Dates) <= 0 {
		return errors.New("dates error")
	}

	if len(task.Seats) <= 0 {
		return errors.New("seats error")
	}

	if len(task.Passengers) <= 0 {
		return errors.New("passengers error")
	}

	// 计算开售时间
	for _, date := range task.Dates {
		var saleTime time.Time
		if saleTime, err = time.Parse("2006-01-02", date); err != nil {
			return errors.New("dates error")
		}

		// 需要减去（预售天数 - 1）
		saleTime = saleTime.Add(from.SaleTime - time.Hour*24*time.Duration(config.Cfg.OtherPresellDays-1))
		t.SaleTime = append(t.SaleTime, saleTime)
	}

	// 开售时间排序
	sort.Slice(t.SaleTime, func(i, j int) bool {
		return t.SaleTime[i].Before(t.SaleTime[j])
	})

	for _, seat := range task.Seats {
		// TODO 转化
		t.Seats = append(t.Seats, seat)
	}

	t.NextQueryTime = time.Now()
	t.TrainNumbers = append(t.TrainNumbers, task.Trains...)
	t.Passengers = append(t.Passengers, task.Passengers...)
	return
}
