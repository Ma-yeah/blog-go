package enums

// ApiModule api 模块

//go:generate stringer -type=Module
type Module int

const (
	FundList Module = iota + 1
	FundDetail
	FundHoliday
	FundDividend
	FundHistory
)
