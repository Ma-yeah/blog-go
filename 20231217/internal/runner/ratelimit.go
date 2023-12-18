package runner

import (
	"20231217/internal/api"
	"context"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"
	"time"
)

type FlowControl struct {
	ctx     context.Context
	limiter *rate.Limiter
}

func NewFlowControl(ctx context.Context) *FlowControl {
	return &FlowControl{
		ctx:     ctx,
		limiter: rate.NewLimiter(rate.Every(time.Second), viper.GetInt("runner.parallel")), // 每秒钟最多 parallel 个请求
	}
}

type RateLimitErr error

// Exec 只返回自身执行中的错误。
func (f *FlowControl) Exec(fapi *api.FundApi) {
	// 等待令牌
	if err := f.limiter.Wait(f.ctx); err != nil {
		fapi.Result.Errorf(RateLimitErr(err))
		return
	}
	// api 的执行错误由 Fn 执行返回，放在 result 中
	fapi.Result = fapi.RequestFn(fapi.Args)
}
