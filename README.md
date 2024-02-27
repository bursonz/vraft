# Raft consensus algorithm

本raft代码实现基于golang语言编写，包含了：

- leader选举
- 日志追加
- 基于内存的持久化
- 以及一些简单优化

实验测试通过自制的模拟网络层进行测试，通讯机制采用了消息报文的形式与原论文中的RPC不同。
测试用例主要借鉴了eliben的检测用例。
模拟网络层通过channel实现、并通过带宽大小模拟传输延迟等网络通讯要素。

TODO:
[] step函数的拆分
[] fakenet 的功能丰富
[] 完全化的消息驱动
[] node耦合功能的剥离

参考：

- https://github.com/eliben/raft
- https://github.com/etcd-io/raft