package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"testing"
	"time"
)

func TestHeartTimeout(t *testing.T) {
	defer func() {
		if exp := recover(); exp != nil {
			fmt.Println(exp)
		}
	}()
	ticker := time.NewTicker(time.Second * 3)
	timer := time.NewTimer(time.Second * 6)
	for {
		select {
		case <-timer.C:
			panic(fmt.Errorf("心跳超时 %v", time.Now()))
		case <-ticker.C:
			fmt.Println("重置心跳", time.Now())
			// 停止定时器后确保 timer 中的 channel 被排空. https://zhuanlan.zhihu.com/p/487913206
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(time.Second * 2)
		}
	}
}

func TestErrGroup(t *testing.T) {
	rand.Seed(time.Now().Unix())
	eg, ctx := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		for {
			tt := time.NewTicker(time.Second * 2)
			select {
			case <-tt.C:
				if rand.Int()%2 == 0 {
					fmt.Println("抛出异常")
					return errors.New("抛出异常")
				}
			}
		}
	})

	eg.Go(func() error {
		tt := time.NewTicker(time.Second * 2)
		for {
			select {
			case <-tt.C:
				fmt.Println("正常输出.")
			case <-ctx.Done():
				fmt.Println("2 收到退出信号")
				return nil
			}
		}
	})

	_ = eg.Wait()
}

func TestTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	fmt.Println("当前时间", time.Now())

	go func(context.Context) {
		time.Sleep(time.Second * 3)
		fmt.Println("打印")
	}(ctx)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("超时结束", time.Now())
			return
		}
	}
}
