package cdn

import (
	"bufio"
	"fmt"
	"gogo12306/logger"
	"gogo12306/worker"
	"net/http"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

func FilterCDN(cdnPath, goodCDNPath string) {
	cdnFile, err := os.Open(cdnPath)
	if err != nil {
		logger.Error("Open CDN file err", zap.String("cdnPath", cdnPath), zap.Error(err))
		return
	}

	goodCDNFile, err := os.Create(goodCDNPath)
	if err != nil {
		logger.Error("Create Available CDN file err", zap.String("goodCDNPath", goodCDNPath), zap.Error(err))
		return
	}

	reader := bufio.NewReader(cdnFile)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	cdnCount := 0
	t0 := time.Now()

	wg := sync.WaitGroup{}
	for scanner.Scan() {
		cdnIP := scanner.Text()

		req, _ := http.NewRequest("POST", fmt.Sprintf("https://%s/otn/index/init", cdnIP), nil)
		req.Host = "kyfw.12306.cn"

		wg.Add(1)
		worker.Do(&worker.Item{
			HttpReq: req,
			Callback: func(body []byte, ok bool, duration time.Duration, err error) {
				// logger.Debug("返回",
				// 	zap.String("ip", cdnIP),
				// 	//   zap.ByteString("body", body),
				// 	zap.Error(err),
				// )

				if ok && len(body) > 0 {
					cdnCount++
					logger.Debug("CDN 可用",
						zap.String("ip", cdnIP),
						// zap.Int("bodyLen", len(body)),
						// zap.ByteString("body", body),
					)

					goodCDNFile.Write([]byte(cdnIP))
					goodCDNFile.Write([]byte("\n"))
				}

				wg.Done()
			},
		})
	}

	wg.Wait()

	logger.Debug("已找到所有可用 CDN", zap.Int("总数", cdnCount), zap.Duration("耗时（秒）", time.Now().Sub(t0)))

	cdnFile.Close()
	goodCDNFile.Close()
}
