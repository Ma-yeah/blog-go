package service

import (
	"20231217/internal/data"
	"20231217/internal/model"
	"context"
	"errors"
	"github.com/bytedance/sonic"
	"github.com/go-redis/redis/v8"
	"time"
)

type Srv struct {
	repo *data.Repo
}

func NewSrv() *Srv {
	return &Srv{
		repo: data.NewRepo(),
	}
}

func (s *Srv) SaveScheduleStatus(ctx context.Context, status *model.ScheduleStatus) error {
	b, err := sonic.Marshal(status)
	if err != nil {
		return err
	}
	_, err = s.repo.Redis.HSet(ctx, "schedule-status", status.Symbol, b).Result()
	return err
}

// GetScheduleStatus 获取品种调度状态
func (s *Srv) GetScheduleStatus(ctx context.Context, symbol string) (*model.ScheduleStatus, error) {
	result, err := s.repo.Redis.HGet(ctx, "schedule-status", symbol).Result()
	if err != nil {
		// 隐藏 redis.Nil 错误
		if errors.Is(err, redis.Nil) {
			err = nil
		}
		return nil, err
	}
	var status *model.ScheduleStatus
	if err = sonic.Unmarshal([]byte(result), status); err != nil {
		return nil, err
	}
	return status, nil
}

func (s *Srv) GetAllFailedSchedule(ctx context.Context) ([]*model.ScheduleStatus, error) {
	result, err := s.repo.Redis.HGetAll(ctx, "schedule-status").Result()
	if err != nil {
		return nil, err
	}
	statuses := make([]*model.ScheduleStatus, 0, len(result))
	for _, v := range result {
		var status *model.ScheduleStatus
		if err = sonic.Unmarshal([]byte(v), status); err != nil {
			return nil, err
		}
		if !status.Success {
			statuses = append(statuses, status)
		}
	}
	return statuses, nil
}

func (s *Srv) GetApikey(ctx context.Context) (string, error) {
	return s.repo.Redis.Get(ctx, "apikey").Result()
}

func (s *Srv) SaveApikey(ctx context.Context, apikey string) error {
	return s.repo.Redis.Set(ctx, "apikey", apikey, time.Minute).Err()
}
