package model

import (
	"20220923/internal/codec"
	"20220923/internal/enum"
	"20220923/internal/typebase"
	"bytes"
	"encoding/binary"
	"fmt"
)

/**
encoding/binary 不能用于编码大小不固定的任意值, 如 int、string
*/

type Message interface {
	Size() uint16
	Type() uint16
}

func NewMessage(msgType uint16) (Message, error) {
	switch msgType {
	case enum.MsgTypeClientDemo:
		return NewClientDemo(), nil
	case enum.MsgTypeServerDemo:
		return NewServerDemo(), nil
	case enum.MsgTypeLogin:
		return NewLogin(), nil
	case enum.MsgTypeLoginResponse:
		return NewLoginResponse(), nil
	case enum.MsgTypeHeartBeat:
		return NewHeartBeat(), nil
	default:
		return nil, fmt.Errorf("unknown MsgType(%d)", msgType)
	}
}

// MetaMessage 消息元数据
type MetaMessage struct {
	MsgSize uint16 `comment:"消息大小"`
	MsgType uint16 `comment:"消息类型"`
}

var MetaMessageSize = binary.Size(MetaMessage{})

func (m MetaMessage) Size() uint16 {
	return m.MsgSize
}

func (m MetaMessage) Type() uint16 {
	return m.MsgType
}

type Login struct {
	MetaMessage

	Username typebase.ByteArrayL12 `comment:"账号(最大支持12位字符)"`
	Password typebase.ByteArrayL20 `comment:"密码(最大支持20位字符)"`
}

var loginMsgSize = uint16(binary.Size(Login{}))

func NewLogin() *Login {
	m := new(Login)
	m.MsgSize = loginMsgSize
	m.MsgType = enum.MsgTypeLogin
	return m
}

type LoginResponse struct {
	MetaMessage

	// Status of the session.
	// 0   - Session Active
	// 5   - Invalid username or IP address
	// 100 - User already connected
	SessionStatus uint8

	_ [3]byte
}

var loginResponseMsgSize = uint16(binary.Size(LoginResponse{}))

func NewLoginResponse() *LoginResponse {
	m := new(LoginResponse)
	m.MsgSize = loginResponseMsgSize
	m.MsgType = enum.MsgTypeLoginResponse
	return m
}

type ClientDemo struct {
	MetaMessage
}

func NewClientDemo() *ClientDemo {
	m := new(ClientDemo)
	m.MsgSize = uint16(MetaMessageSize)
	m.MsgType = enum.MsgTypeClientDemo
	return m
}

type ServerDemo struct {
	MetaMessage
	_      [1]byte
	Addr   typebase.ByteArrayL12
	Port   uint16
	_      [2]byte
	Remark typebase.ByteArrayL20
}

var serverDemoMsgSize = uint16(binary.Size(ServerDemo{}))

func NewServerDemo() *ServerDemo {
	m := new(ServerDemo)
	m.MsgSize = serverDemoMsgSize
	m.MsgType = enum.MsgTypeServerDemo
	return m
}

func (m *ServerDemo) Marshal() ([]byte, error) {
	b := bytes.NewBuffer(make([]byte, 0, serverDemoMsgSize))
	e := codec.NewEncoder(b)
	err := e.Encode(m.MsgSize, m.MsgType, [1]byte{}, m.Addr, m.Port, [2]byte{}, m.Remark)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (m *ServerDemo) Unmarshal(b []byte) error {
	d := codec.NewDecoder(bytes.NewReader(b))
	err := d.Decode(&m.MsgSize, &m.MsgType, &[1]byte{}, &m.Addr, &m.Port, &[2]byte{}, &m.Remark)
	if err != nil {
		return err
	}
	return nil
}

func (m *ServerDemo) String() string {
	str := `{"MsgSize":%d,"MsgType":%d,"Addr":%s,"Port":%d,"Remark":%s}`
	return fmt.Sprintf(str, m.MsgSize, m.MsgType, m.Addr.String(), m.Port, m.Remark.String())
}

type HeartBeat struct {
	MetaMessage
}

func NewHeartBeat() *HeartBeat {
	m := new(HeartBeat)
	m.MsgSize = uint16(MetaMessageSize)
	m.MsgType = enum.MsgTypeHeartBeat
	return m
}
