package api

import (
	"go.uber.org/zap"
	"time"
)

type RequestFunc func(Args) Result
type FundApiErr error

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
	return &FundHistoryResult{Data: []string{"history"}}
}

func fundHistoryDegrade(Args) Result {
	zap.S().Info("fund history degrade")
	time.Sleep(1 * time.Second)
	return &FundHistoryResult{Data: []string{"history"}}
}
