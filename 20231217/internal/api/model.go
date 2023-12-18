package api

import (
	"20231217/internal/enums"
	"time"
)

// Args --->

type Args interface {
	GetSymbol() string
}

type FundApikeyArgs struct {
	AccessKey string
}

func (a *FundApikeyArgs) GetSymbol() string {
	return ""
}

type FundListArgs struct {
	Symbol string
}

func (a *FundListArgs) GetSymbol() string {
	return a.Symbol
}

type FundDetailArgs struct {
	Symbol string
}

func (a *FundDetailArgs) GetSymbol() string {
	return a.Symbol
}

type FundHolidayArgs struct {
	Symbol string
	Year   string
}

func (a *FundHolidayArgs) GetSymbol() string {
	return a.Symbol
}

type FundDividendArgs struct {
	Symbol string
}

func (a *FundDividendArgs) GetSymbol() string {
	return a.Symbol
}

type FundHistoryArgs struct {
	Symbol string
	Start  string
	End    string
}

func (a *FundHistoryArgs) GetSymbol() string {
	return a.Symbol
}

// Result --->

type Result interface {
	Errorf(err error)
	Error() error
	Skipped() bool
}

type FundApikeyResult struct {
	Skip bool // 请求是否被跳过
	Err  FundApiErr
	Data string
}

func (r *FundApikeyResult) Errorf(err error) {
	r.Err = err
}
func (r *FundApikeyResult) Error() error {
	return r.Err
}
func (r *FundApikeyResult) Skipped() bool {
	return r.Skip
}

type FundListResult struct {
	Skip bool // 请求是否被跳过
	Err  FundApiErr
	Data []string
}

func (r *FundListResult) Errorf(err error) {
	r.Err = err
}
func (r *FundListResult) Error() error {
	return r.Err
}
func (r *FundListResult) Skipped() bool {
	return r.Skip
}

type FundDetailResult struct {
	Skip bool // 请求是否被跳过
	Err  FundApiErr
	Data string
}

func (r *FundDetailResult) Errorf(err error) {
	r.Err = err
}
func (r *FundDetailResult) Error() error {
	return r.Err
}
func (r *FundDetailResult) Skipped() bool {
	return r.Skip
}

type FundHolidayResult struct {
	Skip bool // 请求是否被跳过
	Err  FundApiErr
	Data string
}

func (r *FundHolidayResult) Errorf(err error) {
	r.Err = err
}
func (r *FundHolidayResult) Error() error {
	return r.Err
}
func (r *FundHolidayResult) Skipped() bool {
	return r.Skip
}

type FundDividendResult struct {
	Skip bool // 请求是否被跳过
	Err  FundApiErr
	Data string
}

func (r *FundDividendResult) Errorf(err error) {
	r.Err = err
}
func (r *FundDividendResult) Error() error {
	return r.Err
}
func (r *FundDividendResult) Skipped() bool {
	return r.Skip
}

type FundHistoryResult struct {
	Skip bool // 请求是否被跳过
	Err  FundApiErr
	Data []string
}

func (r *FundHistoryResult) Errorf(err error) {
	r.Err = err
}
func (r *FundHistoryResult) Error() error {
	return r.Err
}
func (r *FundHistoryResult) Skipped() bool {
	return r.Skip
}

// FundApi --->

type FundApi struct {
	Module    enums.Module
	Args      Args
	RequestFn RequestFunc
	DegradeFn RequestFunc
	Result    Result
	SkipAble  bool          // 是否是可跳过的 api
	Period    time.Duration // cron 间隔周期
	ApiKey    string
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
		Result:    &FundListResult{Skip: true}, // 因为这个 api 可能被跳过，所以要该给一个默认的 result
		SkipAble:  true,                        // 独立 api，可以跳过的
		Period:    time.Hour,
	}
}

func NewFundDetailApi(Args Args) *FundApi {
	return &FundApi{
		Module:    enums.FundDetail,
		Args:      Args,
		RequestFn: fundDetail,
		DegradeFn: fundDetailDegrade,
		Result:    &FundDetailResult{},
		SkipAble:  false, // 依赖 api，不可跳过的
		Period:    time.Hour * 12,
	}
}

func NewFundHolidayApi(Args Args) *FundApi {
	return &FundApi{
		Module:    enums.FundHoliday,
		Args:      Args,
		RequestFn: fundHoliday,
		Result:    &FundHolidayResult{Skip: true},
		SkipAble:  true, // 独立 api，可以跳过的
		Period:    time.Hour * 24,
	}
}

func NewFundDividendApi(Args Args) *FundApi {
	return &FundApi{
		Module:    enums.FundDividend,
		Args:      Args,
		RequestFn: fundDividend,
		DegradeFn: fundDividendDegrade,
		Result:    &FundDividendResult{},
		SkipAble:  false, // 依赖 api，不可跳过的
		Period:    time.Hour * 24,
	}
}

func NewFundHistoryApi(Args Args) *FundApi {
	return &FundApi{
		Module:    enums.FundHistory,
		Args:      Args,
		RequestFn: fundHistory,
		DegradeFn: fundHistoryDegrade,
		Result:    &FundHistoryResult{},
		SkipAble:  false, // 依赖 api，不可跳过的
		Period:    time.Hour,
	}
}
