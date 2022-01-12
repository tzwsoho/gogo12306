package cdn

import (
	"bufio"
	"fmt"
	"gogo12306/logger"
	"gogo12306/worker"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type CDNInfo struct {
	ResponseTime time.Duration
	IP           string
}

type CDNInfos []*CDNInfo

func (c CDNInfos) Len() int {
	return len(c)
}

func (c CDNInfos) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c CDNInfos) Less(i, j int) bool {
	return c[i].ResponseTime < c[j].ResponseTime
}

func (c CDNInfos) String() string {
	var r []string
	for _, cdn := range c {
		r = append(r, cdn.IP)
	}

	return strings.Join(r, "\n")
}

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

	goodCount := 0
	cdnCount := 0
	t0 := time.Now()

	var cdns CDNInfos
	wg := sync.WaitGroup{}
	for scanner.Scan() {
		cdnIP := scanner.Text()

		req, _ := http.NewRequest("GET", fmt.Sprintf("https://%s/otn", cdnIP), nil)
		req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
		req.Header.Add("Accept-Encoding", "gzip, deflate, br")
		req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
		req.Header.Add("Host", "kyfw.12306.cn")
		req.Header.Add("Connection", "Close")
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.81 Safari/537.36")
		req.Host = "kyfw.12306.cn"

		wg.Add(1)

		t1 := time.Now()
		worker.Do(&worker.Item{
			HttpReq: req,
			Callback: func(body []byte, statusCode int, err error) {
				logger.Debug("返回",
					zap.String("ip", cdnIP),
					zap.Int("bodyLen", len(body)),
					// zap.ByteString("body", body),
					zap.Error(err),
				)

				if statusCode == http.StatusOK || statusCode == http.StatusFound {
					cdnCount++
					duration := time.Since(t1)
					logger.Info("CDN 可用",
						zap.String("ip", cdnIP),
						zap.Duration("耗时", duration),
					)

					if duration < time.Millisecond*300 {
						goodCount++
						cdns = append(cdns, &CDNInfo{
							ResponseTime: duration,
							IP:           cdnIP,
						})
					}
				}

				wg.Done()
			},
		})
	}

	wg.Wait()

	// 按响应时间排序
	sort.Sort(cdns)
	goodCDNFile.Write([]byte(cdns.String()))

	logger.Info("已找到所有可用 CDN",
		zap.Int("总数", cdnCount),
		zap.Int("优秀节点", goodCount),
		zap.Duration("耗时（秒）", time.Since(t0)),
	)

	cdnFile.Close()
	goodCDNFile.Close()
}
