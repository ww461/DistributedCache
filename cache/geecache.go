package cache

import "sync"

//Getter 接口：定义了当缓存未命中时，如何通过 key 从数据源获取原始数据
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//核心调度结构，代表一个缓存命名空间
type Group struct {
	name      string
	getter    Getter //缓存未命中 回源回调
	mainCache cache  //并发安全本地缓存

}
var(
	mu sync.Mutex
	groups = make(map[string]*Group)
)