package lru

//LRU ： Least Recently Used
import "container/list"

//Cache 是一个非并发安全的LRU缓存结构
type Cache struct {
	maxBytes int64
	nBytes   int64
	ll       *list.List               //双向链表
	cache    map[string]*list.Element //key->链表节点的指针

	//可选的回调函数，当某条记录被淘汰时触发(打日志、清理资源)
	OnEvicted func(key string, value Value)
}

//链表中每个节点存的数据
type entry struct {
	key   string
	value Value
}

//任何值只要有长度，内存大小就可以存进缓存
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//1.字典查找节点
//2.移到队首
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

//淘汰最久未访问的节点
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		//更新
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

//Add
//如果key存在，更新值，增加内存差距，移到队首
//不存在，创建新节点，加入字典
//校验是否超出max
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
