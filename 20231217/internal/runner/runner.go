package runner

import (
	"20231217/internal/api"
	"20231217/internal/enums"
	"20231217/internal/model"
	"20231217/internal/service"
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Runner struct {
	cron     *cron.Cron
	ctx      context.Context
	done     chan struct{}     // 用于通知 runner 停止
	notify   chan enums.Module // 用于 cron 到到达时通知
	parallel chan struct{}     // 限制最大可并行执行的任务数量
	wg       sync.WaitGroup
	srv      *service.Srv
	control  *FlowControl
	once     sync.Once
}

func NewRunner(srv *service.Srv) *Runner {
	ctx := context.Background()
	return &Runner{
		cron:     cron.New(),
		ctx:      ctx,
		done:     make(chan struct{}),
		notify:   make(chan enums.Module, 1),
		parallel: make(chan struct{}, viper.GetInt("runner.parallel")), // 这个值不必太大，因为如果出现 apikey 过期，那么失败的请求数量也会变多
		srv:      srv,
		control:  NewFlowControl(ctx),
	}
}

func (r *Runner) Run(subs []model.Subscription) {
	zap.S().Info("runner running...")
reset:
	// 如果之前有未执行完的操作，等待执行完
	r.wg.Wait()
	for i := 0; i < len(subs); {
		select {
		case <-r.done:
			return
		case module := <-r.notify:
			if module == enums.FundList {
				// 请求 fund list
			} else {
				goto reset
			}
		default:
			r.wg.Add(1)
			r.parallel <- struct{}{}
			go func(sub model.Subscription) {
				defer func() {
					<-r.parallel
					r.wg.Done()
				}()
				r.composeRequest(sub)
			}(subs[i])
			i++
		}
	}

	// 检查失败的 api
	for {
		t := time.NewTimer(time.Minute)
		select {
		case <-r.done:
			return
		case module := <-r.notify:
			if module == enums.FundList {
				// 请求 fund list
			} else {
				goto reset
			}
		case <-t.C:
			// 查询所有失败的 api
			status, err := r.srv.GetAllFailedSchedule(r.ctx)
			if err != nil {
				zap.S().Errorf("get all failed schedule error: %v", zap.Error(err))
				continue
			}
			for i := 0; i < len(status); {
				select {
				case <-r.done:
					return
				case module := <-r.notify:
					if module == enums.FundList {
						// 请求 fund list
					} else {
						goto reset
					}
				default:
					r.wg.Add(1)
					r.parallel <- struct{}{}
					go func(sta *model.ScheduleStatus) {
						defer func() {
							<-r.parallel
							r.wg.Done()
						}()
						r.composeRequest()
					}(status[i])
					i++
				}
			}
		}
	}
}

func (r *Runner) request(ctx context.Context, fapi *api.FundApi) api.Result {
	// 获取 api 最近一次的执行状态
	status, err := r.srv.GetScheduleStatus(ctx, fapi.Args.GetSymbol())
	if err != nil {
		fapi.Result.Errorf(err)
		return fapi.Result
	}
	if status != nil {
		// 小于一个定时周期
		if time.Since(time.UnixMilli(status.Ts).UTC()) < fapi.Period {
			// 如果是可跳过的 api，则直接返回
			if fapi.SkipAble {
				// 返回默认的 result，即没有任何数据，调用方也不会有任何操作
				return fapi.Result
			} else {
				// 调用降级方法
				fapi.Result = fapi.DegradeFn(fapi.Args)
				return fapi.Result
			}
		}
	}

	// 必须要获取到 apikey 才能继续
	for {
		key, err := r.srv.GetApikey(r.ctx)
		if err != nil {
			zap.S().Errorf("get apikey from redis error: %v", zap.Error(err))
			// 请求 apikey，需要加分布式锁
			apikey := api.NewFundApikey(&api.FundApikeyArgs{AccessKey: "accesskey"})
			r.control.Exec(apikey)
			if apikey.Result.Error() != nil {
				time.Sleep(time.Second)
			} else {
				key = apikey.Result.(*api.FundApikeyResult).Data
				fapi.ApiKey = key
				if err := r.srv.SaveApikey(r.ctx, key); err != nil {
					zap.S().Errorf("save apikey to redis error: %v", zap.Error(err))
				}
			}
		} else {
			fapi.ApiKey = key
			break
		}
	}

	r.control.Exec(fapi)

	ts := time.Now().UnixMilli()
	if status == nil {
		status = &model.ScheduleStatus{
			Symbol: fapi.Args.GetSymbol(),
			ApiStatus: map[string]model.ApiStatus{
				fapi.Module.String(): {
					Module:  fapi.Module,
					Success: fapi.Result.Error() == nil,
					Args:    fapi.Args,
					Ts:      ts,
				},
			},
			Ts: ts,
		}
	}
	apiStatus, ok := status.ApiStatus[fapi.Module.String()]
	if !ok {
		status.ApiStatus[fapi.Module.String()] = model.ApiStatus{
			Module:  fapi.Module,
			Success: fapi.Result.Error() == nil,
			Args:    fapi.Args,
			Ts:      ts,
		}
	}
	apiStatus.Ts = ts
	apiStatus.Args = fapi.Args
	apiStatus.Success = fapi.Result.Error() == nil

	if fapi.Result.Error() != nil {
		status.Success = false
	} else {
		// 记录调度成功到 redis
		if len(status.ApiStatus) != 4 { // 除去 apikey 和 list 的 api
			status.Success = false
		} else {
			status.Success = true
			for _, v := range status.ApiStatus {
				if !v.Success {
					status.Success = false
					break
				}
			}
		}
	}

	status.ApiStatus[fapi.Module.String()] = apiStatus
	if err := r.srv.SaveScheduleStatus(r.ctx, status); err != nil {
		zap.S().Errorf("save schedule status to redis error: %v", zap.Error(err))
	}
	return fapi.Result
}

// Compose 对 api 请求按顺序组合
func (r *Runner) composeRequest(sub model.Subscription) {
	defer func() {
		// 应该在这里记录 api 掉色状态，而不是在 request 中
		// 不要异步，意义不大，反而会变麻烦
	}()
	// holiday 是独立的
	holidayResult := r.request(r.ctx, api.NewFundHolidayApi(&api.FundHolidayArgs{Symbol: sub.Symbol}))
	if holidayResult.Error() != nil {
		// holiday 失败则记录，不影响其他 api
		zap.S().Error("holiday", zap.Error(holidayResult.Error()))
	} else if !holidayResult.Skipped() {
		fmt.Println(holidayResult)
	}

	// detail 是 dividend 的前置
	detailResult := r.request(r.ctx, api.NewFundDetailApi(&api.FundDetailArgs{Symbol: sub.Symbol}))
	if detailResult.Error() != nil {
		// 记录调度状态到 redis
		return
	}
	fmt.Println(detailResult)
	// dividend 是 history 的前置
	dividendResult := r.request(r.ctx, api.NewFundDividendApi(&api.FundDividendArgs{Symbol: sub.Symbol}))
	if dividendResult.Error() != nil {
		// 记录调度状态到 redis
		return
	}
	fmt.Println(dividendResult)
	// history 要分多次请求
	historyResult := r.request(r.ctx, api.NewFundHistoryApi(&api.FundHistoryArgs{Symbol: sub.Symbol, Start: "2021-01-01", End: "2021-01-02"}))
	if historyResult.Error() != nil {
		// 记录调度状态到 redis
		return
	}
	fmt.Println(historyResult)
}

func (r *Runner) Stop() {
	zap.S().Info("runner stopping...")
	close(r.done)
	r.cron.Stop()
	close(r.parallel)
	r.wg.Wait()
	zap.S().Info("runner stopped")
}
