README
===========================
远程调用(RPC)
-----------
#### 要求
* 设计一个分布式处理系统。一个节点（进程）作为Master,负责节点注册，任务分配；其它多个节点（进程）作为worker，执行客户端发送的请求。系统执行流程为：
* Master首先启动
* Worker节点启动后向Master进行“注册”；Master维护所有已注册节点的信息（比如IP地址和TCP/UDP端口，服务类型等）；
* 客户端提交任务前首先向Master发送“查询”消息，询问将任务提交给哪个worker
* Master根据一定策略（如轮询）从所有Worker中选择一个worker，比如A，将其信息发送给Client
* Worker A完成任务后将结果返回给客户端
* 计算任务为：两个数的+，-，*，/。数据类型支持32位整数、64位整数、浮点数
* 要求消息编码格式：google protocol buffer
* Worker端要支持多个客户端请求的并发处理
* 测试要求：
    * Worker数量不少于3个
    * 客户端数量不少于3个
    * 单个客户端发送请求数要大于100万个，或者测试时间不少于10分钟
