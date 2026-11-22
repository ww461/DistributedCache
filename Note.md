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