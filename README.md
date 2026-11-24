# DistributedCache - 从零实现的轻量级分布式内存缓存

基于 Go 语言从零实现的轻量级分布式缓存系统（参考 Google `groupcache` 架构），支持 LRU 缓存淘汰、一致性哈希节点路由、Singleflight 防缓存击穿及 Protobuf 高性能节点间通信。

---

## 🌟 核心特性与开发路线图

- [x] **Day 1: LRU 核心缓存与只读保护**
  - 基于双向链表 + Map 实现 $O(1)$ 复杂度的 LRU 缓存淘汰策略
  - 引入 `ByteView` 只读字节视图提供深拷贝保护，防止缓存外部污染
- [x] **Day 2: 单机并发缓存与 Getter 回调**
  - 使用互斥锁（`sync.Mutex`）保证缓存并发安全
  - 采用接口型函数设计模式实现灵活的数据源回源逻辑
- [ ] **Day 3: HTTP 服务端节点**
  - 基于 `net/http` 搭建支持分布式通信的 HTTPPool 节点
- [ ] **Day 4: 一致性哈希算法（Consistent Hashing）**
  - 实现虚拟节点机制，解决节点增删时的数据倾斜与大规模缓存失效
- [ ] **Day 5: 分布式节点通信**
  - 抽象节点通信接口，打通跨节点缓存自动路由与拉取
- [ ] **Day 6: Singleflight 防止缓存击穿**
  - 并发控制合并重复请求，避免瞬时高并发击穿底层数据库
- [ ] **Day 7: Protobuf 协议优化**
  - 节点间通信全面接入 Protobuf 二进制序列化

---


## 🚀 快速开始

### 环境依赖
- Go 1.22+

### 运行单元测试
```bash
# 测试 Day 1 LRU 模块
go test -v ./cache/lru