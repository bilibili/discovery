## 背景

* 团队是B站主站技术，位于上海的Gopher
* 之前一直使用zookeeper作为服务注册发现中间件，但由于zookeeper：
    * 用好难，维护难，二次开发难
    * 服务规模太大(3W+节点)，一套zk集群很难满足大量的连接和写请求，但如果拆分服务类型维护多套集群又破坏了连通性
    * CP系统，对于微服务的服务注册发现，其实不如一套AP系统更可用
* 运维小锅锅们大力推进业务向k8s的迁移，zk难以满足迭代需求
* 也考虑过直接使用现有的如consul、etcd，但我们更想要一套纯纯的AP系统
    * 补充资料：[Why You Shouldn’t Use ZooKeeper for Service Discovery](https://medium.com/knerd/eureka-why-you-shouldnt-use-zookeeper-for-service-discovery-4932c5c7e764)

## 设计目标

Netflix Eureka可以说是服务注册发现领域AP系统的标杆，所以参考Eureka之后，我们设定以下目标：

1. 实现AP类型服务注册发现系统，在可用性极极极极强的情况下，努力保证数据最终一致性
2. 与公司k8s平台深度结合，注册打通、发布平滑、naming service等等
3. 网络闪断等异常情况，可自我保护，保证每个节点可用
4. 基于HTTP协议实现接口，简单易用，维护各流行语言SDK

## 相对Netflix Eureka的改进

* 长轮询监听应用变更（Eureka定期30s拉取一次）
* 只拉取感兴趣的AppID实例（Eureka一拉就是全部，无法区分）
* 合并node之间的同步请求/(ㄒoㄒ)/~~其实还没实现，是个TODO
* Dashboard骚操作~
* 多注册中心信息同步支持
* 更完善的日志记录

PS：[Eureka2.0](https://github.com/Netflix/eureka/wiki/Eureka-2.0-Motivations)以上功能基本也要实现，但难产很久了/(ㄒoㄒ)/~~
