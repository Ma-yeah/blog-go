package main

import (
	"20220923/internal/enum"
	_ "20220923/internal/log"
	"20220923/internal/model"
	"20220923/internal/packet"
	"context"
	"errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {
	server := newTCPServer()
	if err := server.Start(); err != nil {
		panic(err)
	}

	shutdown := make(chan os.Signal, 1)
	// 监听 os.Interrupt, os.Kill 这两种信号，一但收到信号，则将信号写入 shutdown 中
	signal.Notify(shutdown, os.Interrupt, os.Kill)
	<-shutdown

	_ = server.Stop()
}

func handler(conn *net.TCPConn) {
	// 对于服务端来说，remote 表示客户端地址，local 表示服务端地址
	zap.S().Debugf("accepted. client_addr=[%s], server_adrr=[%s]", conn.RemoteAddr().(*net.TCPAddr).IP.To4().String(), conn.LocalAddr().(*net.TCPAddr).IP.To4().String())
	// 包装
	s := packet.NewSession(conn)
	// 错误恢复
	defer func() {
		if exp := recover(); exp != nil {
			zap.S().Errorf("[recovered] panic: %v\n", exp)
		}
		_ = s.Close()
		zap.S().Warn("disconnect with client")
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 登录请求
	loginPkt, err := s.ReadPacket(ctx)
	if err != nil {
		panic(err)
	}
	loginMsg, err := loginPkt.ReadMessage()
	if err != nil {
		panic(err)
	}
	loginResp := model.NewLoginResponse()
	if loginMsg.Type() != enum.MsgTypeLogin {
		loginResp.SessionStatus = 5
		loginRespPkt := new(packet.Buffer)
		err = loginRespPkt.WriteMessage(loginResp)
		if err != nil {
			panic(err)
		}
		err = s.WritePacket(ctx, loginRespPkt)
		if err != nil {
			panic(err)
		}
		return
	}
	login := loginMsg.(*model.Login)
	uname := login.Username.String()
	passw := login.Password.String()
	if uname != serverUsername || passw != serverPassword {
		loginResp.SessionStatus = 5
		loginRespPkt := new(packet.Buffer)
		err = loginRespPkt.WriteMessage(loginResp)
		err = s.WritePacket(ctx, loginRespPkt)
		if err != nil {
			panic(err)
		}
		return
	} else {
		loginResp.SessionStatus = 0
		loginRespPkt := new(packet.Buffer)
		err = loginRespPkt.WriteMessage(loginResp)
		err = s.WritePacket(ctx, loginRespPkt)
		if err != nil {
			panic(err)
		}
	}
	// errgroup 是 waitgroup 的一个包装，同样实现了一个 goroutine 等待多个 goroutine，同时可以返回 error
	// errgroup 中自动创建了 context.WithCancel。当任一个 eg.Go 中 return 了 error，会执行 ctx 的 cancel，return nil 则不不会执行 cancel，但他们都会调用内部的 wg.Done()
	eg, ctx := errgroup.WithContext(ctx)
	ch := make(chan *packet.Buffer)
	heart := make(chan struct{})

	// 读取客户端消息
	eg.Go(func() error {
		for {
			pkt, err := s.ReadPacket(ctx)
			if err != nil {
				return err
			}
			msg, err := pkt.ReadMessage()
			if err != nil {
				return err
			}
			switch msg.Type() {
			case enum.MsgTypeLogin:
				loginResp = model.NewLoginResponse()
				loginResp.SessionStatus = 100
				p := new(packet.Buffer)
				_ = p.WriteMessage(loginResp)
				ch <- p
			case enum.MsgTypeHeartBeat:
				// 写入一个任意值
				heart <- struct{}{}
			case enum.MsgTypeClientDemo:
				// 响应数据
				m := model.NewServerDemo()
				m.Addr.Set([]byte("127.0.0.1"))
				m.Port = serverPort
				m.Remark.Set([]byte("server response: ok"))
				p := new(packet.Buffer)
				err = p.WriteMessage(m)
				if err != nil {
					return err
				}
				ch <- p
			}
		}
	})

	// 维持心跳
	eg.Go(func() error {
		// 心跳间隔
		ticker := time.NewTicker(time.Second * 5)
		// 大于 3 倍心跳间隔
		timer := time.NewTimer(time.Second * 17)
		m := model.NewHeartBeat()
		for {
			select {
			case <-ticker.C:
				p := new(packet.Buffer)
				_ = p.WriteMessage(m)
				ch <- p
			case <-heart:
				zap.S().Debug("receive client heartBeat")
				// 重置心跳超时
				if !timer.Stop() {
					select {
					case <-timer.C: // 确保定时器中的 channel 被排空
					default:
					}
				}
				timer.Reset(time.Second * 17)
			case <-timer.C:
				return errors.New("client heart timeout")
			case <-ctx.Done():
				return nil
			}
		}
	})

	// 写回客户端消息
	eg.Go(func() error {
		for {
			select {
			case p := <-ch:
				err = s.WritePacket(ctx, p)
				if err != nil {
					return err
				}
			case <-ctx.Done():
				return nil
			}
		}
	})

	if err := eg.Wait(); err != nil {
		panic(err)
	}
}

type TCPServer struct {
	lis *net.TCPListener

	userName string
	password string
}

const (
	serverUsername = "mayee"
	serverPassword = "mayee"
	serverPort     = 30001
)

func newTCPServer() *TCPServer {
	// 如果未指定 IP，则自动选择本地一个可用的 unicast and anycast 地址监听；如果未指定 Port(默认 0)，则随机选一个端口监听
	lis, _ := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: serverPort})
	return &TCPServer{
		lis:      lis,
		userName: serverUsername,
		password: serverPassword,
	}
}

func (s *TCPServer) Start() error {
	for {
		conn, err := s.lis.AcceptTCP()
		if err != nil {
			return err
		}
		go handler(conn)
	}
}

func (s *TCPServer) Stop() error {
	return s.lis.Close()
}
