package cdn

import (
	"bufio"
	"gogo12306/logger"
	"math/rand"
	"os"

	"go.uber.org/zap"
)

var cdns []string

func init() {
	cdns = make([]string, 0)
}

func LoadCDN(goodCDNPath string) (err error) {
	var fCDN *os.File
	if fCDN, err = os.Open(goodCDNPath); err != nil {
		logger.Warn("打开可用 CDN 文件失败，建议先用 -c 参数启动本程序筛选可用 CDN 列表，当前默认使用 kyfw.12306.cn", zap.String("cdnPath", goodCDNPath), zap.Error(err))
		return nil
	}

	reader := bufio.NewReader(fCDN)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		ip := scanner.Text()
		cdns = append(cdns, ip)
	}

	fCDN.Close()

	logger.Info("可用 CDN 数量", zap.Int("count", len(cdns)))
	return nil
}

func GetCDN() string {
	return "kyfw.12306.cn"
}

func GetCDN0() string {
	if len(cdns) == 0 {
		return "kyfw.12306.cn"
	}

	// return cdns[rand.Intn(len(cdns)*10000)/10000]

	n := len(cdns)
	if n > 10 {
		n = 10
	}

	return cdns[rand.Intn(n*10000)/10000]
}
