package typebase

type ByteArrayL12 [12]byte

func (s *ByteArrayL12) Set(b []byte) {
	copy(s[:], b)
	// 位数不足时补 0
	for i := len(b); i < len(s); i++ {
		s[i] = 0
	}
}

func (s *ByteArrayL12) Bytes() []byte {
	for i := len(s); i > 0; i-- {
		// 32 在 ASCII 码中表示一个 空格
		if s[i-1] != 0 && s[i-1] != 32 {
			// 把末端的 0 填充或 空格 去掉
			return s[:i]
		}
	}
	// 返回一个空字节数组
	return s[:0]
}

func (s *ByteArrayL12) String() string {
	return string(s.Bytes())
}

type ByteArrayL20 [20]byte

func (s *ByteArrayL20) Set(b []byte) {
	copy(s[:], b)
	// 位数不足时补 0
	for i := len(b); i < len(s); i++ {
		s[i] = 0
	}
}

func (s *ByteArrayL20) Bytes() []byte {
	for i := len(s); i > 0; i-- {
		// 32 在 ASCII 码中表示一个 空格
		if s[i-1] != 0 && s[i-1] != 32 {
			// 把末端的 0 或 空格 去掉
			return s[:i]
		}
	}
	// 返回一个空字节数组
	return s[:0]
}

func (s *ByteArrayL20) String() string {
	return string(s.Bytes())
}
