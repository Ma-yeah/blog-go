package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"time"
)

func TestBinarySize(t *testing.T) {
	var a uint8 = 10
	var b uint16 = 10
	var c uint32 = 10
	var d uint64 = 10
	println(binary.Size(a)) // 1
	println(binary.Size(b)) // 2
	println(binary.Size(c)) // 4
	println(binary.Size(d)) // 8
}

func TestByteBuffer(t *testing.T) {
	buffer := bytes.Buffer{}
	buffer.Write([]byte{1, 2, 3, 4, 5})
	fmt.Printf("%v\n", buffer.Bytes())
	b := make([]byte, 2)
	_, _ = buffer.Read(b) // 从 byte 中读取走了字节
	fmt.Printf("%v\n", b)
	next := buffer.Next(2) // 从 byte 中读取走了字节
	fmt.Printf("%v\n", next)
	fmt.Printf("%v\n", buffer.Bytes())
}

func TestChannel(t *testing.T) {
	ch := make(chan int)

	go func(ch chan int) {
		time.Sleep(time.Second * 3)
		ch <- 1
	}(ch)

	for {
		select {
		case i := <-ch:
			fmt.Println(i)
			return
		}
	}
}
