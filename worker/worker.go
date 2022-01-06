package worker

import (
	"gogo12306/httpcli"
	"gogo12306/logger"
	"net/http/cookiejar"
	"time"

	"go.uber.org/zap"
)

const (
	GOROUTINE_MAX = 200
	INTERVAL      = time.Second * 3
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

func DoTask(jar *cookiejar.Jar, task *Task) {
	go func(j *cookiejar.Jar, t *Task) {
		tk := time.NewTicker(time.Second)
		for {
			select {
			case <-t.Done:
				return

			case <-tk.C:
				now := time.Now()

				if now.Before(t.NextQueryTime) { // 未到查询时间
					// 重新设置下次查询时间
					delta := time.Millisecond * 100

					sub := now.Sub(t.NextQueryTime)
					if sub > time.Minute {
						delta = time.Minute

						logger.Info("未到查询时间",
							zap.String("任务下次开始时间", now.Add(delta).String()),
						)
					} else if sub > INTERVAL {
						delta = INTERVAL

						logger.Info("未到查询时间",
							zap.String("任务下次开始时间", now.Add(delta).String()),
						)
					}

					tk.Reset(delta)
					break
				}

				// 判断当前时间是否在 12306 开放的 6~23 点之间，不在的话定时到明天 5 点 59 分开始
				// if now.Hour() < 6 || now.Hour() >= 23 {
				// 	_, zone := now.Zone()
				// 	delta := now.Truncate(time.Hour*24).AddDate(0, 0, 1).Add(time.Hour*6-time.Hour*time.Duration(zone/3600)).Sub(now) - time.Minute
				// 	tk.Reset(delta)

				// 	logger.Warn("不在 12306 营业时间内",
				// 		zap.String("任务开始时间", now.Add(delta).String()),
				// 	)

				// 	break
				// }

				// 判断是否已到最早的开售时间，不在的话定时到开售前 1 分钟开始
				if now.Before(t.SaleTimes[0]) {
					delta := t.SaleTimes[0].Sub(now) - time.Minute
					tk.Reset(delta)

					logger.Warn("未到开售时间",
						zap.String("开售时间", t.SaleTimes[0].String()),
						zap.String("任务开始时间", now.Add(delta).String()),
					)

					break
				}

				logger.Info("任务开始",
					zap.String("出发站", task.From),
					zap.String("到达站", task.To),
					zap.Strings("出发日期", task.StartDates),
				)

				// go t.CB(j, t)
				t.CB(j, t)

				tk.Reset(INTERVAL)
			}
		}
	}(jar, task)
}
