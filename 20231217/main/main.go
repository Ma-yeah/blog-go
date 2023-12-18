package main

import (
	"20231217/internal/model"
	"20231217/internal/runner"
	"20231217/internal/service"
	"github.com/spf13/viper"
	"time"
)

// 模拟外不配置的环境变量
func init() {
	// redis 地址
	viper.SetDefault("data.redis.addrs", []string{"127.0.0.1:6379"})
	// 表示并行执行的品种数量
	viper.SetDefault("runner.parallel", 5)
}

func main() {
	// 订阅品种,
	subs := []model.Subscription{
		{Symbol: "MOCK_001"},
		{Symbol: "MOCK_002"},
		{Symbol: "MOCK_003"},
		{Symbol: "MOCK_004"},
		{Symbol: "MOCK_005"},
		{Symbol: "MOCK_006"},
		{Symbol: "MOCK_007"},
		{Symbol: "MOCK_008"},
	}

	r := runner.NewRunner(service.NewSrv())
	r.Run(subs)
	// 模拟执行一段时间后，外部停止执行
	time.Sleep(time.Minute * 1)
	r.Stop()
}
