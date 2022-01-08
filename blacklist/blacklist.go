package blacklist

import (
	"fmt"
	"gogo12306/config"
	"sync"
	"time"
)

// 小黑屋
// Key: 任务 ID + "_" + CDN IP + "_" + 车次 + "_" + 席位类型
// Value: 解除时间
var blackList sync.Map

func makeKey(taskID int64, trainCode string, seatIndex int) string {
	return fmt.Sprintf("%d_%s_%d", taskID, trainCode, seatIndex)
}

func AddToBlackList(taskID int64, trainCode string, seatIndex int) {
	blackList.Store(makeKey(taskID, trainCode, seatIndex),
		time.Now().Add(time.Duration(config.Cfg.Login.BlackTime)*time.Second))
}

func IsInBlackList(taskID int64, trainCode string, seatIndex int) bool {
	key := makeKey(taskID, trainCode, seatIndex)
	if t, exists := blackList.Load(key); exists {
		if tt, ok := t.(time.Time); ok {
			if time.Now().Before(tt) {
				return true
			}
		}
	}

	blackList.Delete(key)
	return false
}
