package cache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

// 模拟一个底层的慢速数据库
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 测试 1：测试 GetterFunc 接口型函数是否能正常转换
func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

// 测试 2：测试缓存未命中时回源，以及第二次查询时是否命中缓存
func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++ // 每次去数据库查就加 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range db {
		// 第 1 次获取：缓存没有，调用 GetterFunc 回源查 db
		view, err := gee.Get(k)
		if err != nil || view.String() != v {
			t.Fatal("failed to get value of " + k)
		}
		// 第 2 次获取：此时缓存已经有了，应该直接命中缓存，loadCounts 不增加！
		_, err = gee.Get(k)
		if err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("the value of unknown should be empty, but %s got", view)
	}
}
