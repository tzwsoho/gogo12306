package worker

import (
	"gogo12306/httpcli"
	"gogo12306/logger"
	"time"

	"go.uber.org/zap"
)

const (
	GOROUTINE_MAX = 50
	INTERVAL_MAX  = time.Minute
	INTERVAL_MIN  = time.Second * 3
)

var (
	items          chan *Item
	itemGoRoutines int
)

func init() {
	items = make(chan *Item, GOROUTINE_MAX)
}

func Do(item *Item) {
	items <- item

	if itemGoRoutines >= GOROUTINE_MAX {
		return
	}

	go func(n int) {
		for item := range items {
			if item.Callback != nil {
				item.Callback(httpcli.DoHttp(item.HttpReq, nil))
			} else {
				httpcli.DoHttp(item.HttpReq, nil)
			}
		}
	}(itemGoRoutines)

	itemGoRoutines++
}

func DoTask(task *Task) {
	go func(t *Task) {
		tk := time.NewTicker(time.Nanosecond)
		for range tk.C {
			now := time.Now()
			// logger.Debug("Now", zap.String("now", now.String()))

			if now.Before(t.NextQueryTime) { // 未到查询时间
				// 重新设置下次判断时间
				delta := time.Millisecond * 100

				sub := now.Sub(t.NextQueryTime)
				if sub > INTERVAL_MAX {
					delta = INTERVAL_MAX
				} else if sub > INTERVAL_MIN {
					delta = INTERVAL_MIN
				}

				tk.Reset(delta)

				logger.Debug("任务下次开始时间为", zap.String("时间", now.Add(delta).String()))
				continue
			}

			// 判断当前时间是否在 12306 开放的 6~23 点之间，不在的话定时到明天 5 点 59 分开始
			if now.Hour() < 6 || now.Hour() >= 23 {
				_, zone := now.Zone()
				delta := now.Truncate(time.Hour*24).AddDate(0, 0, 1).Add(time.Hour*6-time.Hour*time.Duration(zone/3600)).Sub(now) - time.Minute
				tk.Reset(delta)

				logger.Debug("任务下次开始时间为", zap.String("时间", now.Add(delta).String()))
				continue
			}

			// 判断是否已到开售时间，不在的话定时到开售前 1 分钟开始
			if now.Before(t.SaleTime[0]) {
				delta := t.SaleTime[0].Sub(now) - time.Minute
				tk.Reset(delta)

				logger.Debug("任务下次开始时间为", zap.String("sale", t.SaleTime[0].String()), zap.String("时间", now.Add(delta).String()))
				continue
			}

			logger.Debug("Begin", zap.String("now", now.String()))

			tk.Reset(INTERVAL_MIN)
		}
	}(task)
}
