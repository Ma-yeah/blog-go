package api

import (
	"context"
	"golang.org/x/time/rate"
	"time"
)

type flowControl struct {
	ctx     context.Context
	limiter *rate.Limiter
}

var global *flowControl

func init() {
	global = &flowControl{
		ctx:     context.Background(),
		limiter: rate.NewLimiter(rate.Every(time.Second), 5), // 每秒钟最多 5 个请求
	}
}

type RateLimitErr error

// Exec 只返回自身执行中的错误。
func Exec(api *FundApi) RateLimitErr {
	// 等待令牌
	if err := global.limiter.Wait(global.ctx); err != nil {
		return RateLimitErr(err)
	}
	// api 的执行错误由 Fn 执行返回，放在 result 中
	api.Result = api.RequestFn(api.Args)
	return nil
}
