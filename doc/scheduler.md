## 简介
多机房部署时，需要对各个机房的流量进行控制调度，多机房流量调度主要解决一下几个问题

1. 根据监控获取的机房负载指标，动态下发机房负载调度信息，是个机房流量尽量负载均衡
2. 机房故障时，通过下发机房调度信息，动态调度流量到可用机房，保证流量的平稳切换

## 实现
多机房流量调度的实现包括

1. discovery下发负载调度信息
2. sdk根据负载信息进行流量调度

## discovery 实现

* 通过poll / fetch 接口动态返回多机房的流量调度信息

调度信息格式如下(scheduler 部分)
```
  {
    "code":0,
    "message":"0",
    "ttl":1,
    "data":{
        "main.account.account-service":{
            "instances":[....],
            "scheduler":[{"src":"sh001","dst":{"sh001":3,"sh002":1}},{"src":"bj002","dst":{"bj001":1,"bj002":3}}],
            "latest_timestamp":1545795463166919001,
            "latest_timestamp_str":"1545795463"
        },
        "latest_timestamp":1525948297987066600
    }
}
``` 
* scheduler 返回信息说明

|参数	|含义|
|--|--|
|scheduler.src	|源机房信息|
|scheduler.dst|	服务提供方所在机房信息|
|dst.key	|服务提供方所在目标机房|
|dst.value| 服务提供方所在机房的流量权重|
上面示例的返回信息表示:
* sh001 的服务调度3/4的流量到sh001 ，调度1/4的流量到sh002机房
* bj002 的服务调度 1/4 的流量到bj001 , 调度3/4的流量到bj002 机房

## sdk实现
* sdk使用二级调度进行流量的负载调度

1. 根据返回的scheduler判断自身所处的机房匹配src获取机房调度信息

如：部署在sh001的服务匹配到到 
```
{
  "src":"sh001",
  "dst":{
  "sh001":3,
  "sh002":1
  }
},
```
则需要调度3/4的流量到sh001机房，1/4的流量到sh002 机房

2. 根据机房获取机房应用实例，并由实例的weight进行节点负载的二次调度
参考 [sdk 实现](https://github.com/bilibili/discovery/blob/master/naming/naming.go#L76)

 

 