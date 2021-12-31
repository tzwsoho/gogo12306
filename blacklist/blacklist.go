package blacklist

import (
	"fmt"
	"sync"
	"time"
)

const BlackTime = time.Minute

// 小黑屋
// Key: 任务 ID + "_" + CDN IP + "_" + 车次 + "_" + 席位类型
// Value: 解除时间
var blackList sync.Map

func makeKey(taskID int64, trainCode string, seatIndex int) string {
	return fmt.Sprintf("%d_%s_%d", taskID, trainCode, seatIndex)
}

func AddToBlackList(taskID int64, trainCode string, seatIndex int) {
	blackList.Store(makeKey(taskID, trainCode, seatIndex), time.Now().Add(BlackTime))
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
