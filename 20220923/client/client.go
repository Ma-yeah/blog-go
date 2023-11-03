package main

import (
	"20220923/internal/enum"
	_ "20220923/internal/log"
	"20220923/internal/model"
	"20220923/internal/packet"
	"context"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net"
	"time"
)

func main() {
	// 指定使用本地的哪个 ip 发起连接
	local := &net.TCPAddr{IP: net.ParseIP("127.0.0.1")}
	// 指定连接到哪个 ip 和 端口
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 30001}
	// 这么做的意义在于：当客户机有多块网卡，其中一个网卡的 ip 在服务端的白名单中，如果客户端未指定 ip，则可能会用到另一个非白名单的 ip 发起连接，导致被服务端拒绝
	// 测试方法：客户机电脑同时连上网线和WLAN，这样就有两个网卡地址。然后分别用这两个地址发起 tcp 连接，在服务端观察打印出来的客户端地址
	conn, err := net.DialTCP("tcp", local, remote)
	if err != nil {
		panic(err)
	}
	s := packet.NewSession(conn)
	defer func() {
		if exp := recover(); exp != nil {
			zap.S().Errorf("[recovered] panic: %v\n", exp)
		}
		_ = s.Close()
		zap.S().Warn("disconnect with server")
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 登录
	login := model.NewLogin()
	login.Username.Set([]byte("mayee"))
	login.Password.Set([]byte("mayee"))
	loginPkt := new(packet.Buffer)
	err = loginPkt.WriteMessage(login)
	if err != nil {
		panic(err)
	}
	err = s.WritePacket(ctx, loginPkt)
	if err != nil {
		panic(err)
	}
	loginRespPkt, err := s.ReadPacket(ctx)
	if err != nil {
		panic(err)
	}
	message, err := loginRespPkt.ReadMessage()
	if err != nil {
		panic(err)
	}
	if message.Type() != enum.MsgTypeLoginResponse {
		panic("login response type error")
	}
	response := message.(*model.LoginResponse)
	if response.SessionStatus != 0 {
		panic(fmt.Errorf("lgoin response %d", response.SessionStatus))
	}

	eg, ctx := errgroup.WithContext(ctx)
	ch := make(chan *packet.Buffer)

	// 接收消息
	eg.Go(func() error {
		for {
			pkt, err := s.ReadPacket(ctx)
			if err != nil {
				return err
			}
			m, err := pkt.ReadMessage()
			if err != nil {
				return err
			}
			switch m.Type() {
			case enum.MsgTypeHeartBeat:
				zap.S().Debug("receive server heartBeat")
				// 响应心跳
				// 必须要重新写入一次 message, 因为 p.ReadMessage 后，p.num 为 0 了。若此时发送 p 后，服务端调用 p.ReadMessage, 发现 p.num 为 0 则手动抛出 io.EOF
				_ = pkt.WriteMessage(m)
				ch <- pkt
			case enum.MsgTypeServerDemo:
				demo := m.(*model.ServerDemo)
				zap.S().Info(demo.String())
			}
		}
	})

	// 发送消息
	eg.Go(func() error {
		demo := model.NewClientDemo()
		t := time.NewTicker(time.Second * 2)
		for {
			select {
			case <-t.C:
				p := new(packet.Buffer)
				_ = p.WriteMessage(demo)
				err = s.WritePacket(ctx, p)
				if err != nil {
					panic(err)
				}
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

	if err = eg.Wait(); err != nil {
		panic(err)
	}
}

func netDemo() {
	// Linux 可以使用`ip a`命令查看网卡索引和名称；Windows 可以使用`netsh int ipv4 show interfaces`查看网卡索引和名称
	// 列出所有的网卡
	itfs, _ := net.Interfaces()
	for _, itf := range itfs {
		// 有了索引和名称可以通过 net.InterfaceByIndex() 或 net.InterfaceByName() 获取指定网卡
		fmt.Printf("网卡索引=[%d], 网卡名称=[%s]\n", itf.Index, itf.Name)
		addrs, _ := itf.Addrs()
		for _, addr := range addrs {
			if ip := addr.(*net.IPNet).IP.To4(); ip != nil {
				fmt.Printf("ip 地址: %s", ip.String())
				break
			}
		}
	}
}
