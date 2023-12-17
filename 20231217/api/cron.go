package api

import (
	"20231217/enums"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

type CronNotify struct {
	// cron 触发信号
	Trigger chan enums.Module
}

var Notify = &CronNotify{Trigger: make(chan enums.Module, 1)}

// 初始化定时任务
func init() {
	c := cron.New()
	_, _ = c.AddFunc(viper.GetString("fund.cron.list"), func() {
		Notify.Trigger <- enums.FundList
	})
	_, _ = c.AddFunc(viper.GetString("fund.cron.detail"), func() {
		Notify.Trigger <- enums.FundDetail
	})
	_, _ = c.AddFunc(viper.GetString("fund.cron.holiday"), func() {
		Notify.Trigger <- enums.FundHoliday
	})
	_, _ = c.AddFunc(viper.GetString("fund.cron.dividend"), func() {
		Notify.Trigger <- enums.FundDividend
	})
	_, _ = c.AddFunc(viper.GetString("fund.cron.history"), func() {
		Notify.Trigger <- enums.FundHistory
	})
}
