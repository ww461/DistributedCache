package cache

// 只有一个只读属性的b，持有缓存的真实字节数据
type ByteView struct {
	b []byte
}

//返回缓存所占内存大小
func (v ByteView) Len() int {
	return len(v.b)
}

//深拷贝切片
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

//返回拷贝副本
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

//将字节以字符串形式返回，打印，调试
func (v ByteView) String() string {
	return string(v.b)
}
