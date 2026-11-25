package consistenthash

//自定义哈希函数
type Hash func(data []byte) uint32

type Map struct{
	hash Hash
	replicas int //一个真实节点对应的虚拟节点个数
	keys []int //哈希环
	hashMap map[int]string //虚拟节点与真实节点的映射表，key = 虚拟节点的哈希值，value = 真实节点名称
	
}