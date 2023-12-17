package main

import (
	"20231217/api"
	"20231217/enums"
	"context"
	"fmt"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	// 限制最大可并行请求的 api 数量 （建议是并行请求 ifast 接口的 2 倍）
	parallel := make(chan struct{}, 10)
	symbols := subscribeSymbols()
	// 并发级别在品种之间，单个品种内部串行请求

	// apikey 不能进调度，只能是手动请求的
reset:
	for i := 0; i < len(symbols); {
		select {
		case module := <-api.Notify.Trigger:
			// 如果是基金列表触发，则请求
			if module == enums.FundList {
				if result, err := api.NewFundListApi(api.FundListArgs{}).Request(ctx); err != nil {
					zap.S().Error("request fund list error: ", err.Error())
				} else {
					// 存储
					fmt.Println(result)
				}
			} else {
				goto reset
			}
		default:
			parallel <- struct{}{}
			func(symbol string) {
				defer func() { <-parallel }()
				api.Compose(ctx, symbol)
			}(symbols[i])
			i++
		}
	}
}

// 获取订阅品种
func subscribeSymbols() []string {
	return []string{"USDT", "BTC", "ETH"}
}
