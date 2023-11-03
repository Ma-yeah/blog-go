package typebase

import (
	"fmt"
	"testing"
)

func TestByte(t *testing.T) {
	b := []byte("hello")
	fmt.Printf("%v\n", b) // 转成了对应的 ASCII 码值
}
