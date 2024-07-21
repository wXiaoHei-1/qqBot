package utils

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"math"
	constant "qqbot/constant"
	"time"
)

// ParseData 解析数据
func ParseData(message []byte, target interface{}) error {
	data := gjson.Get(string(message), "d")
	return json.Unmarshal([]byte(data.String()), target)
}

// GetOpMeans OPMeans 返回 op 含义
func GetOpMeans(op int) string {
	means, ok := constant.OpMeans[op]
	if !ok {
		means = "unknown"
	}
	return means
}

// CTW concurrencyTimeWindowSec 并发时间窗口，单位秒
const CTW = 2

// CalcInterval 根据并发要求，计算连接启动间隔
func CalcInterval(maxC uint32) time.Duration {
	if maxC == 0 {
		maxC = 1
	}
	f := math.Round(CTW / float64(maxC))
	if f == 0 {
		f = 1
	}
	return time.Duration(f) * time.Second
}
