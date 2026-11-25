package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

//自定义哈希函数
type Hash func(data []byte) uint32

//一致性哈希的核心数据结构
type Map struct {
	hash     Hash           //哈希函数
	replicas int            //一个真实节点对应的虚拟节点个数
	keys     []int          //哈希环：存放所有虚拟节点的哈希值
	hashMap  map[int]string //虚拟节点与真实节点的映射表，key = 虚拟节点的哈希值，value = 真实节点名称
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil{
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

//向哈希环里面添加真实节点
func (m *Map) Add(keys ...string){
	for _,key := range keys{
		//创建replicas个虚拟节点
		for i := 0;i < m.replicas;i++{
			//虚拟节点名称 = 编号 + 真实节点名
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	//排序方便后续查找
	sort.Ints(m.keys)
}

//Get
// 根据给定的key在哈希环上顺时针找距离最近的节点，返回真实节点
func(m *Map) Get(key string) string{
	if len(m.keys) == 0{
		return ""
	}

	//计算key的哈希值
	hash := int(m.hash([]byte(key)))

	//找到数据挂载的服务器的虚拟节点
	idx := sort.Search(len(m.keys),func(i int)bool{
		return m.keys[i] >= hash
	})

	//if key > max(keys) sort.Search return len(m.keys)
	//through % len(m.keys) return idx = 0
	return m.hashMap[m.keys[idx % len(m.keys)]]
}