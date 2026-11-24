// 实现核心的HTTP处理逻辑，实现go标准库的http.handler接口
package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecache"

// 承载节点间HTTP通信的核心服务
type HTTPPool struct {
	self     string //自己的地址
	basePath string //节点通信的前缀 _/geecache
}

// 初始化服务池
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// 打印带有节点身份标识的日志
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// 实现handler接口，负责拦截并处理所有外部发来的HTTP缓存请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//检查:路径以basePath开头
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexppected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	//约定访问格式/<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request:path must be /<basepath>/<groupname>/<key>", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	//找缓存组
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	//group.Get(key) 捞取数据
	view,err := group.Get(key)
	if err != nil{
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}

	//将缓存字节数据写入响应体返回
	w.Header().Set("Content-Type","application/octet-stream")
	w.Write(view.ByteSlice())
}
