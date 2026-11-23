package cache

import (
	"DistributedCache/cache/lru"
	"sync"
)

//对LRU添加互斥锁，提供本地并发安全
//1.Mutex互斥锁保护内部的lru.Cache
//2.懒加载 调用add时才去初始化lru

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64 //缓存最大运行字节数
}

// add 并发安全添加缓存
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	//
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

// get
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
