package model

import (
	"20231217/internal/api"
	"20231217/internal/enums"
)

// Subscription 订阅品种
type Subscription struct {
	Symbol string `json:"symbol"`
}

// ScheduleStatus 记录调度状态
type ScheduleStatus struct {
	Symbol    string               `json:"symbol"`
	ApiStatus map[string]ApiStatus `json:"api_status"`
	Success   bool                 `json:"success"` // 如果 compose 中的所有的 api 都成功，则 success 为 true
	Ts        int64                `json:"ts"`      // 最近一次调度完成的时间（注意，内部 api 如果重试了，不改变这个时间）。当下依次调度到来时，不判断内部 api 的状态
}

// ApiStatus 记录 api 执行状态
type ApiStatus struct {
	Module  enums.Module `json:"module"`
	Args    api.Args     `json:"args"`
	Success bool         `json:"success"` // 表示这个 api 最近一次是否执行成功
	Ts      int64        `json:"ts"`      // api 最近执行的时间。如果发生了重试，则此时间更新
}
