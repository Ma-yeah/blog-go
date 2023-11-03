package packet

import (
	"20220923/internal/codec"
	"20220923/internal/model"
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"
)

type Session struct {
	net.Conn
	*Reader
	*Writer
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		Conn:   conn,
		Reader: NewReader(conn),
		Writer: NewWriter(conn),
	}
}

var headerSize = uint16(binary.Size(header{}))

// binary_size = 2 + 1 + 1 + 4 + 8 = 16
type header struct {
	PktSize  uint16 `comment:"数据包大小"`
	MsgCount uint8  `comment:"数据包中消息的个数"`
	_        byte   `comment:"无实际意义"`
	SeqNum   uint32 `comment:"数据包中第一条消息的序号"`
	SendTime uint64 `comment:"数据包发送时间(毫秒级时间戳)"`
}

type Buffer struct {
	header `comment:"数据包头"`

	mu  sync.RWMutex `comment:"读写锁"`
	buf bytes.Buffer `comment:"数据包体(消息内容)"`
	num uint8        `comment:"消息数量"`
}

type Reader struct {
	mu sync.Mutex
	rd *bufio.Reader
}

type Writer struct {
	mu sync.Mutex
	wr *bufio.Writer
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		rd: bufio.NewReader(r),
	}
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		wr: bufio.NewWriter(w),
	}
}

// WritePacket 写入数据包
func (w *Writer) WritePacket(ctx context.Context, p *Buffer) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	w.mu.Lock()
	defer w.mu.Unlock()

	eh := make(chan error)
	done := make(chan struct{})

	go func() {
		h := new(header)
		// 数据包的大小 = 包头 + 包体
		h.PktSize = headerSize + uint16(p.buf.Len())
		h.MsgCount = p.num
		h.SendTime = uint64(time.Now().UnixMilli())
		if p.SeqNum != 0 {
			h.SeqNum = p.SeqNum
		}
		b, err := codec.Marshal(h)
		if err != nil {
			eh <- err
			return
		}
		// 写入包头
		if _, err = w.wr.Write(b); err != nil {
			eh <- err
			return
		}
		// 写入包体
		if _, err = p.buf.WriteTo(w.wr); err != nil {
			eh <- err
			return
		}
		// 发出数据
		if err = w.wr.Flush(); err != nil {
			eh <- err
			return
		}
		// 重置序号，因为 p 可能会被重复使用
		p.num = 0
		done <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-eh:
			return err
		case <-done:
			return nil
		}
	}
}

// ReadPacket 读出数据包
func (r *Reader) ReadPacket(ctx context.Context) (*Buffer, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	bh := make(chan *Buffer)
	eh := make(chan error)
	go func() {
		// 用于接受数据包头
		b := make([]byte, headerSize)
		if _, err := io.ReadFull(r.rd, b); err != nil {
			eh <- err
			return
		}
		p := new(Buffer)
		if err := codec.Unmarshal(b, &p.header); err != nil {
			eh <- err
			return
		}
		// 用于接受数据包体
		b = make([]byte, p.PktSize-headerSize)
		if _, err := io.ReadFull(r.rd, b); err != nil {
			eh <- err
			return
		}
		p.buf.Write(b)
		p.num = p.MsgCount
		bh <- p
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-eh:
			return nil, err
		case b := <-bh:
			return b, nil
		}
	}
}

// WriteMessage 写入一个消息 PacketChan 的包装
func (p *Buffer) WriteMessage(m model.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	b, err := codec.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := p.buf.Write(b); err != nil {
		return err
	}
	p.num++
	return nil
}

// ReadMessage 读出一个消息
func (p *Buffer) ReadMessage() (model.Message, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	h := new(model.MetaMessage)
	// 没有消息，或者消息长度不完整
	if p.num == 0 || p.buf.Len() < model.MetaMessageSize {
		return nil, io.EOF
	}
	// 查看消息头长度的字节，返回了一个副本并反序列化。此时，p.buf 中的字节并没有被取出
	if err := codec.Unmarshal(p.buf.Bytes()[:model.MetaMessageSize], h); err != nil {
		return nil, err
	}
	// 消息不完整
	if p.buf.Len() < int(h.MsgSize) {
		return nil, io.EOF
	}
	// 根据消息类型创建出结构体
	m, err := model.NewMessage(h.MsgType)
	if err != nil {
		return nil, err
	}
	// 从 p.buf 中读出 MsgSize 个字节并反序列。等同于 Next 方法等同于 Read，只是不会返回 error
	if err := codec.Unmarshal(p.buf.Next(int(h.MsgSize)), m); err != nil {
		return nil, err
	}
	p.num--
	return m, nil
}

func (p *Buffer) FirstMessage() (model.Message, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.num == 0 {
		return nil, io.EOF
	}
	h := new(model.MetaMessage)
	if err := codec.Unmarshal(p.buf.Bytes()[:model.MetaMessageSize], h); err != nil {
		return nil, err
	}
	m, err := model.NewMessage(h.MsgType)
	if err != nil {
		return nil, err
	}
	if err := codec.Unmarshal(p.buf.Bytes()[:h.MsgSize], m); err != nil {
		return nil, err
	}
	return m, nil
}

func (p *Buffer) LastMessage() (model.Message, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.num == 0 {
		return nil, io.EOF
	}
	h := new(model.MetaMessage)
	var offset int
	for i := uint8(0); i < p.num; i++ {
		if err := codec.Unmarshal(p.buf.Bytes()[offset:offset+model.MetaMessageSize], h); err != nil {
			return nil, err
		}
		offset += int(h.MsgSize)
	}
	m, err := model.NewMessage(h.MsgType)
	if err != nil {
		return nil, err
	}
	if err := codec.Unmarshal(p.buf.Bytes()[offset-int(h.MsgSize):], m); err != nil {
		return nil, err
	}
	return m, nil
}
