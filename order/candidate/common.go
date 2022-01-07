package candidate

import (
	"net/http/cookiejar"

	"gogo12306/common"
	"gogo12306/worker"
)

func DoCandidate(jar *cookiejar.Jar, task *worker.Task, leftTicketInfo *common.LeftTicketInfo,
	seatIndex int) (err error) {
	if err = CheckFace(jar, &CheckFaceRequest{
		SecretStr: leftTicketInfo.SecretStr,
		SeatIndex: seatIndex,
	}); err != nil {
		return
	}

	return
}
