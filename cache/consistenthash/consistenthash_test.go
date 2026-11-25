package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	// 自定义一个确定性的哈希函数：直接把字节数字转为整数，方便人工推算结果
	// 设虚拟节点倍数为 3
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// 添加 3 个真实节点："6", "4", "2"
	// 生成的虚拟节点在环上的位置分别是：
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2", // 2  -> 刚好命中 2  -> 真实节点 "2"
		"11": "2", // 11 -> 顺时针最近是 12 -> 真实节点 "2"
		"23": "4", // 23 -> 顺时针最近是 24 -> 真实节点 "4"
		"27": "2", // 27 -> 比最大的 26 还大，绕一圈回到起点 2 -> 真实节点 "2"
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	// 动态扩容：新增真实节点 "8"，新增虚拟节点：8, 18, 28
	hash.Add("8")

	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
}