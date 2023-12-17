package model

import (
	"20231217/api"
	"20231217/data"
	"20231217/enums"
	"context"
	"errors"
	"github.com/bytedance/sonic"
	"github.com/go-redis/redis/v8"
)

// ScheduleStatus 记录调度状态
type ScheduleStatus struct {
	Symbol    string      `json:"symbol"`
	ApiStatus []ApiStatus `json:"api_status"`
	Success   bool        `json:"success"` // 如果 compose 中的所有的 api 都成功，则 success 为 true
	Ts        int64       `json:"ts"`      // 最近一次调度完成的时间（注意，内部 api 如果重试了，不改变这个时间）。当下依次调度到来时，不判断内部 api 的状态
}

// ApiStatus 记录 api 执行状态
type ApiStatus struct {
	Module  enums.Module `json:"module"`
	Args    api.Args     `json:"args"`
	Success bool         `json:"success"` // 表示这个 api 最近一次是否执行成功
	Ts      int64        `json:"ts"`      // api 最近执行的时间。如果发生了重试，则此时间更新
}

// SaveScheduleStatus 保存品种的调度状态
func SaveScheduleStatus(ctx context.Context, status *ScheduleStatus) error {
	b, err := sonic.Marshal(status)
	if err != nil {
		return err
	}
	_, err = data.GlobalRedis.HSet(ctx, "schedule-status", status.Symbol, b).Result()
	return err
}

// GetScheduleStatus 获取品种调度状态
func GetScheduleStatus(ctx context.Context, symbol string) (*ScheduleStatus, error) {
	result, err := data.GlobalRedis.HGet(ctx, "schedule-status", symbol).Result()
	if err != nil {
		// 隐藏 redis.Nil 错误
		if errors.Is(err, redis.Nil) {
			err = nil
		}
		return nil, err
	}
	var status *ScheduleStatus
	if err = sonic.Unmarshal([]byte(result), status); err != nil {
		return nil, err
	}
	return status, nil
}
