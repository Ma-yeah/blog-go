package api

import (
	"20231217/enums"
	"20231217/model"
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"
)

type Args interface {
	GetSymbol() string
}

type args struct {
	Symbol string
}

func (a *args) GetSymbol() string {
	return a.Symbol
}

type FundApikey struct {
	Args
	AccessKey string
}

type FundListArgs struct {
	Args
}

type FundDetailArgs struct {
	*args
}

type FundHolidayArgs struct {
	*args
	Year string
}

type FundDividendArgs struct {
	*args
}

type FundHistoryArgs struct {
	*args
	Start string
	End   string
}

type Result interface {
	Error() error
}

type result struct {
	Err error
}

func (r *result) Error() error {
	return r.Err
}

type FundApikeyResult struct {
	*result
	Data string
}

type FundListResult struct {
	*result
	Data []string
}

type FundDetailResult struct {
	*result
	Data string
}

type FundHolidayResult struct {
	*result
	Data string
}

type FundDividendResult struct {
	*result
	Data string
}

type FundHistoryResult struct {
	*result
	Data string
}

type RequestFunc func(Args) Result

var (
	_ RequestFunc = fundApikey
	_ RequestFunc = fundList
	_ RequestFunc = fundDetail
	_ RequestFunc = fundHoliday
	_ RequestFunc = fundDividend
	_ RequestFunc = fundHistory
)

func fundApikey(Args) Result {
	time.Sleep(1 * time.Second)
	return &FundApikeyResult{Data: "apikey"}
}

// FundList 历史列表
func fundList(Args) Result {
	time.Sleep(1 * time.Second)
	return &FundListResult{Data: []string{"a", "b", "c"}}
}

// FundDetail 基金详情
func fundDetail(Args) Result {
	time.Sleep(1 * time.Second)
	return &FundDetailResult{Data: "detail"}
}

func fundDetailDegrade(Args) Result {
	zap.S().Info("fund detail degrade")
	time.Sleep(1 * time.Second)
	return &FundDetailResult{Data: "detail"}
}

// FundHoliday 基金节假日
func fundHoliday(Args) Result {
	time.Sleep(1 * time.Second)
	return &FundHolidayResult{Data: "holiday"}
}

// FundDividend 基金节假日
func fundDividend(Args) Result {
	time.Sleep(1 * time.Second)
	return &FundDividendResult{Data: "dividend"}
}

func fundDividendDegrade(Args) Result {
	zap.S().Info("fund dividend degrade")
	time.Sleep(1 * time.Second)
	return &FundDividendResult{Data: "dividend"}
}

// FundHistory 基金历史净值
func fundHistory(Args) Result {
	time.Sleep(1 * time.Second)
	return &FundHistoryResult{Data: "history"}
}

func fundHistoryDegrade(Args) Result {
	zap.S().Info("fund history degrade")
	time.Sleep(1 * time.Second)
	return &FundHistoryResult{Data: "history"}
}

// Compose 对 api 请求按顺序顺序组合
func Compose(ctx context.Context, symbol string) {
	// holiday 是独立的
	holidayResult, err := NewFundHolidayApi(FundHolidayArgs{args: &args{Symbol: symbol}, Year: "2023"}).Request(ctx)
	if err != nil || holidayResult.Error() != nil {
		// holiday 失败则记录，不影响其他 api
		zap.S().Error("holiday", zap.Error(holidayResult.Error()))
	} else {
		fmt.Println(holidayResult)
	}

	// detail 是 dividend 的前置
	detailResult, err := NewFundDetailApi(FundDetailArgs{args: &args{Symbol: symbol}}).Request(ctx)
	if err != nil || detailResult.Error() != nil {
		// 记录调度状态到 redis
		return
	}
	fmt.Println(detailResult)
	// dividend 是 history 的前置
	dividendResult, err := NewFundDividendApi(FundDividendArgs{args: &args{Symbol: symbol}}).Request(ctx)
	if err != nil || dividendResult.Error() != nil {
		// 记录调度状态到 redis
		return
	}
	fmt.Println(dividendResult)
	// history 要分多次请求
	historyResult, err := NewFundHistoryApi(FundHistoryArgs{args: &args{Symbol: symbol}, Start: "2022-01-01", End: "2023-01-31"}).Request(ctx)
	if err != nil || historyResult.Error() != nil {
		// 记录调度状态到 redis
		return
	}
	fmt.Println(historyResult)
}

type FundApi struct {
	Module    enums.Module
	Args      Args
	RequestFn RequestFunc
	DegradeFn RequestFunc
	Result    Result
	SkipAble  bool // 是否是可跳过的 api
	ApiKey    string
}

func (api *FundApi) Request(ctx context.Context) (Result, error) {
	// 获取 api 最近一次的执行状态
	status, err := model.GetScheduleStatus(ctx, api.Args.GetSymbol())
	if err != nil {
		return nil, err
	}
	if status != nil {
		// 小于一个定时周期
		if time.Since(time.UnixMilli(status.Ts).UTC()) < time.Hour {
			// 如果是可跳过的 api，则直接返回
			if api.SkipAble {
				// 返回默认的 result，即没有任何数据，调用方也不会有任何操作
				return api.Result, nil
			} else {
				// 调用降级方法
				api.Result = api.DegradeFn(api.Args)
				return api.Result, nil
			}
		}
	}

	// 必须要获取到 apikey 才能继续
	for {
		key := GetApiKey(ctx)
		if key != "" {
			api.ApiKey = key
			break
		}
		// 请求 apikey
		apikey := NewFundApikey(FundApikey{AccessKey: "accesskey"})
		if rateLimitErr := Exec(apikey); rateLimitErr == nil && apikey.Result.Error() == nil {
			api.ApiKey = apikey.Result.(*FundApikeyResult).Data
		}
		time.Sleep(time.Second)
	}

	if err = Exec(api); err != nil || api.Result.Error() != nil {
		// 记录调度失败到 redis
	} else {
		// 记录调度成功到 redis
	}
	return api.Result, err
}

func NewFundApikey(Args) *FundApi {
	return &FundApi{
		RequestFn: fundApikey,
	}
}

func NewFundListApi(Args Args) *FundApi {
	return &FundApi{
		Module:    enums.FundList,
		Args:      Args,
		RequestFn: fundList,
		Result:    FundListResult{result: &result{Err: nil}}, // 因为这个 api 可能被跳过，所以要该给一个默认的 result
		SkipAble:  true,                                      // 独立 api，可以跳过的
	}
}

func NewFundDetailApi(Args Args) *FundApi {
	return &FundApi{
		Module:    enums.FundDetail,
		Args:      Args,
		RequestFn: fundDetail,
		DegradeFn: fundDetailDegrade,
		SkipAble:  false, // 依赖 api，不可跳过的
	}
}

func NewFundHolidayApi(Args Args) *FundApi {
	return &FundApi{
		Module:    enums.FundHoliday,
		Args:      Args,
		RequestFn: fundHoliday,
		Result:    FundHolidayResult{result: &result{Err: nil}},
		SkipAble:  true, // 独立 api，可以跳过的
	}
}

func NewFundDividendApi(Args Args) *FundApi {
	return &FundApi{
		Module:    enums.FundDividend,
		Args:      Args,
		RequestFn: fundDividend,
		DegradeFn: fundDividendDegrade,
		SkipAble:  false, // 依赖 api，不可跳过的
	}
}

func NewFundHistoryApi(Args Args) *FundApi {
	return &FundApi{
		Module:    enums.FundHistory,
		Args:      Args,
		RequestFn: fundHistory,
		DegradeFn: fundHistoryDegrade,
		SkipAble:  false, // 依赖 api，不可跳过的
	}
}

func GetApiKey(ctx context.Context) string {
	return "apikey"
}
