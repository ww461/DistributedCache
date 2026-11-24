// 实现核心的HTTP处理逻辑，实现go标准库的http.handler接口
package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecache"

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

//实现handler接口，负责拦截并处理所有外部发来的HTTP缓存请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter,r *http.Request){
	//检查
	if !strings.HasPrefix(r.URL.Path,p.basePath){
		panic("HTTPPool serving unexppected path: " + r.URL.Path)
	}
	p.Log("%s %s",r.Method,r.URL.Path)
}