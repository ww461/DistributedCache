# DistributedCache

# Day1
LRU 缓存核心
单机内存最基础的那块缓存
内存大小有限，存满时淘汰一部分旧数据
数据结构选择：双向链表+ 哈希表
    哈希表实现O（1）根据key查找数据
双向链表 go自带的container/list
    维护访问顺序,每访问一个元素就移到队首；当内存超出阈值，淘汰队尾元素，O（1）

## byteview.go 
- 封装 缓存值 []byte
缓存可以存储文本、图片、序列化的二进制对象，用[]byte 存储通用
b []byte
Len() int
- 为什么不直接调用len
只作用于内置类型
b是私有成员，包外拿不到，不能让拿到
为了实现Value接口，解耦
    lru 是一个通用的缓存淘汰算法包，它在设计时根本不知道、也不想知道你未来存的是什么类型。  
    你可能存我们的 ByteView，也可能存自定义的 String，甚至未来存一个 Image 结构体。
    lru 只提了一个要求：“不管你是猫是狗，只要你能告诉我你占多大地方（实现了 Len() int 方法），你就可以存进我的 LRU 缓存！”
    只要 ByteView 拥有了 Len() int 方法，它就自动满足了 Value 接口（Go 的鸭子类型），从而能被放进 LRU 中
len和内存的关系
    []byte 长度就是字节数，字节数就是内存大小
byte 是一个数据类型，和int一模一样的性质
`type byte = uint8`
byte 是8位（1字节）的无符号整数 存储0-255整数，可以应对ASCII
内存寻址的最小基本单位是字节,所有切片的长度就说内存的大小
ByteView是自定义的结构体
ByteSlice() []byte 返回b的深拷贝副本
为什么不直接暴露字节切片，go的切片底层是引用传递，外部可以修改，缓存会被污染
，只需一个只读视图提供深拷贝保护
## lru
lru.go
- 实现LRU的Cache结构体以及Get，Add，RemoveOldest
Cache 结构体
├── ll: 双向链表 (*list.List)          --> 维护访问顺序（队首最新，队尾最老）
├── cache: 哈希表 (map[string]*list.Element) --> 实现 O(1) 快速定位链表节点
├── maxBytes: 最大允许占用内存 (int64)
└── nbytes: 当前已使用的总内存 (int64)
- ll       *list.List 没有数据类型
go的标准库的空接口设计，类型断言
```
       list.Element (节点结构体)
┌─────────────────────────────────┐
│ prev: 指向前一个节点的指针      │
│ next: 指向后一个节点的指针      │
│                                 │
│ Value (interface{} 字段):       │
│     │                           │
│     ▼                           │
│   &entry{                       │
│       key:   "Tom",             │
│       value: String("630"),     │
│   }                             │
└─────────────────────────────────┘

}
```
interface{} any 可以装进任何类型的数据
业务层人为规定只存*entry结构体指针

- kv := ele.Value.(*entry)
go的双向链表存空接口，任何类型
list.Element是一个结构体表示节点元素，里面有个字段是Value才是节点值
本身传进list就是 *entry 转化为any
.(*entry) 将空接口转化为 *entry类型
==断言就是知道接口本身是什么类型然后进行转化==
- nbytes 存的是kv字节数

## test
go test 
go的单元测试机制
当你运行 go test ./cache/lru 时，Go 命令行工具在后台悄悄帮你做了这几件事：
自动搜寻：扫描 ./cache/lru 目录，找出所有以 _test.go 结尾的文件。
提取测试函数：找出里面所有名字形如 func TestXxx(t *testing.T) 的函数。
偷偷生成一个临时 main 程序：Go 会在后台自动生成一个包含 main() 的临时程序，里面去逐个调用这些 TestXxx 函数。
编译并运行：把这个临时程序编译运行，收集每个测试的结果（看有没有触发 t.Fatalf 报错），输出运行耗时和结果，最后自动把临时文件删掉。
你不需要自己写任何 main() 入口，Go 工具链帮你全自动搞定了。
文件名必须以 _test.go 结尾
函数签名必须符合固定格式
func TestXxx(t *testing.T) {
    // 函数名必须以大写 "Test" 开头，后面接大写字母（如 TestGet, TestRemoveoldest）
    // 参数必须是 t *testing.T
}
t.Fatalf("...")：打印错误信息并立即终止当前测试。
t.Errorf("...")：打印错误信息，但继续往下执行其他测试。

func main() {
    lru := New(...)
    lru.Add("a", "1")
    fmt.Println(lru.Get("a")) // 人肉用眼睛看屏幕上打印出来的东西对不对
}在实际工程中，这种方式有很大弊端： 
无法自动化：容易遗漏，污染入口，单元测试的好处是：
写好测试断言（比如 if lru.Len() != 2 { t.Fatalf(...) }），以后无论你怎么重构底层算法，只要在终端敲一句 go test ./...，0.1 秒内就能全自动告诉你所有逻辑是否依然正常！

# Day2
- lru的map是不支持并发读写的，10个goroutine去读写map，抛出崩溃 fatal error : concurrent map read and map write
需要加互斥锁
sync.Mutex
- 区分lru.Cache 和 cache.cache 
一个是底层算法实现
一个是封装+并发
lru.maxbytes = 底层能用多少
lru.nbytes = 现在用了多少
cache.cachebytes = 外层为了懒加载而暂存的上限
```
Group (NewGroup 传入 cacheBytes: 100MB)
  │
  ▼
cache 结构体 (cache.go)
  ├── cacheBytes = 100MB   <── 【暂存配置】：因为 lru 还未创建，先把它记在身上
  └── lru *lru.Cache       <── 此时还是 nil (懒加载)
        │
        │ 当第一次调用 add() 时，执行: lru.New(c.cacheBytes, nil)
        ▼
      lru.Cache 结构体 (lru.go)
        ├── maxBytes = 100MB  <── 【真正的上限】：底层算法根据它决定何时淘汰
        └── nbytes   = 45KB   <── 【当前使用量】：随着 Add/RemoveOldest 实时增减
```
- add 为什么不传Value 而是 BytesView
lru 包是底层的“通用算法工具”，而 cache 包是具体的“分布式缓存系统”。
    1.网络传输的数据必须是纯粹的二进制字节流（[]byte）。
    ByteView 内部存的就是 b []byte，天生具备网络传输和序列化能力；
    但如果传 lru.Value，它只是一个宽泛的接口（只要求有 Len() int），外部如果塞一个结构体进去，网络层根本不知道该怎么通过网络把它发给另一台服务器。  
    2. 暴露Value

1. 原始数据 (一堆数字)
   []byte  -> 比如 [54, 51, 48] (代表 "630")
     │
     ▼  (第 1 层包装：装入安全盒子)
2. ByteView 结构体
   ByteView{ b: []byte }
   // 这个盒子很特殊：它天生自带一个方法 func (v ByteView) Len() int
     │
     ▼  (第 2 层隐式转换：装入接口大箱子)
3. 存入 lru.Cache 时，转成了 Value 接口！
   func (c *Cache) Add(key string, value Value)
     │
     ▼  (最终存入双向链表)
4. entry 结构体
   type entry struct {
       key   string
       value Value // 这个 Value 接口，肚子里装着真实的 ByteView
   }
- 将 v->byteView

## 互斥锁的作用
Go 语言会直接强制崩溃 Go 官方为了性能，默认的 map 是非并发安全的。如果协程 A 正在给 map 写入新数据，协程 B 刚好在读取 map，Go 的底层运行时（Runtime）会直接抛出致命错误：
双向链表指针错乱 两个协程同时往双向链表头插节点，可能导致 node.prev 和 node.next 指针相互覆盖，最后链表形成死循环环路，CPU 瞬间飙到 100%，或者内存泄漏。
- 任何人想进去查缓存或者写缓存，先执行 c.mu.Lock()

# geecache.go
核心门面 Group，协调本地缓存与数据回源。
缓存未命中时去数据源捞数据，写回缓存
每个业务方的数据源五花八门（有的查 MySQL，有的查 Redis，有的调用第三方 HTTP API），我们的缓存库怎么做到通用？
解决办法：让调用方传一个“回源回调函数”。这里会用到 Go 语言中极其经典且优雅的“接口型函数设计模式”

```
type Getter interface {
    Get(key string) ([]byte, error)
}
type GetterFunc func(key string) ([]byte, error)
func (f GetterFunc) Get(key string) ([]byte, error) {
    return f(key)
}
```
- 实现了Get 你就算时Getter
- 定义一个函数类型GetterFunc 哦哦哦 一个函数类型是接收string 返回[]byte error 这个函数就可以说是Getterfunc
-  给getterfunc类型定义一个Get方法，直接调用函数本身 这样Getterfunc也是Getter
- 设计是给NewGroup用的
