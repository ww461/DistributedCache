package cache

import (
	"fmt"
	"log"
	"sync"
)

// Getter 接口：定义了当缓存未命中时，如何通过 key 从数据源获取原始数据
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// 核心调度结构，代表一个缓存命名空间
type Group struct {
	name      string
	getter    Getter //缓存未命中 回源回调
	mainCache cache  //并发安全本地缓存

}

var (
	mu     sync.Mutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.Lock()
	defer mu.Unlock()

	g := groups[name]
	return g
}

// 从缓存中获取对应key的值
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	//缓存里面没有,load 回源
	return g.load(key)
}

// load加载数据
func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

// 从本地回源
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) //数据库拿(外部)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key,value)
	return value,nil
}

func (g *Group) populateCache(key string,value ByteView){
	g.mainCache.add(key,value)
}
