- [字段定义](#字段定义)
- [错误码定义](#错误码定义)
- [注册register](#注册register)
- [心跳renew](#心跳renew)
- [下线cancel](#下线cancel)
- [获取实例fetch](#获取实例fetch)
- [批量获取实例fetchs](#批量获取实例fetchs)
- [长轮询获取实例poll](#长轮询获取实例poll)
- [长轮询批量获取实例polls](#长轮询批量获取实例polls)
- [获取node节点](#获取node节点)
- [修改实例信息set](#修改实例信息set)


### 字段定义

| 字段     | 说明                                                                                                                             |
| -------- | -------------------------------------------------------------------------------------------------------------------------------- |
| zone     | 机房服务地区标识，用于多机房部署区分数据中心                                                                                     |
| env      | 环境信息，(例如：fat1,uat ,pre ,prod)分别对应fat环境 集成环境，预发布和线上                                                          |
| appid    | 服务唯一标识。【业务标识.服务标识[.子服务标识]】 全局唯一，禁止修改                                                            |
| hostname | instance主机标识                                                                                                            |
| addrs    | 服务地址 格式为 scheme://ip:port，支持多个协议地址。如 grpc://127.0.0.1:8888, http://127.0.0.1:8887                             |
| color    | 服务标记，可用于集群区分，业务灰度流量选择集群                                                                                   |
| version  | 服务版本号信息                                                                                                                 |
| metadata | 服务自定义扩展元数据，格式为{"key1":"value1"}，可以用于传递权重，负载等信息 使用json格式传递。  { “weight":"10","key2":"value2"} |

### 错误码定义ecode

| 错误码 | 说明           |
| ------ | -------------- |
| 0      | 成功           |
| -304   | 实例信息无变化 |
| -400   | 请求参数错误   |
| -404   | 实例不存在     |
| -409   | 实例信息不一致 |
| -500   | 未知错误       |

### 注册register

*HTTP*

POST http://HOST/discovery/register

*请求参数*

| 参数名   | 必选  | 类型              | 说明                             |
| -------- | ----- | ----------------- | -------------------------------- |
| zone     | true  | string            | 可用区                           |
| env      | true  | string            | 环境                             |
| appid    | true  | string            | 服务名标识                       |
| hostname | true  | string            | 主机名                           |
| addrs    | true  | []string          | 服务地址列表                     |
| status   | true  | int               | 状态，1表示接收流量，2表示不接收 |
| color    | false | string            | 灰度或集群标识                   |
| metadata | false | json string | 业务自定义信息      必须为map[string]string 的json格式            |

*返回结果*

```json
*****成功*****
{
    "code":0,
    "message":""
}
****失败****
{
    "code":-400,
    "message":"-400"
}
```

*CURL*
```shell
curl 'http://127.0.0.1:7171/discovery/register' -d "zone=sh1&env=test&appid=provider&hostname=myhostname&status=1&addrs=http%3A%2F%2F172.1.1.1%3A8000&addrs=grpc%3A%2F%2F172.1.1.1%3A9999&version=111&metadata=%7B%22weight%22%3A10%7D"
```

### 心跳renew

*HTTP*

POST http://HOST/discovery/renew

*请求参数*

| 参数名   | 必选  | 类型              | 说明                             |
| -------- | ----- | ----------------- | -------------------------------- |
| zone     | true  | string            | 可用区                           |
| env      | true  | string            | 环境                             |
| appid    | true  | string            | 服务名标识                       |
| hostname | true  | string            | 主机名                           |

*返回结果*

```json
*****成功*****
{
    "code":0,
    "message":""
}
****失败****
{
    "code":-400,
    "message":"-400"
}
```

*CURL*
```shell
curl 'http://127.0.0.1:7171/discovery/renew' -d "zone=sh1&env=test&appid=provider&hostname=myhostname"
```

### 下线cancel

*HTTP*

POST http://HOST/discovery/cancel

*请求参数*

| 参数名   | 必选  | 类型              | 说明                             |
| -------- | ----- | ----------------- | -------------------------------- |
| zone     | true  | string            | 可用区                           |
| env      | true  | string            | 环境                             |
| appid    | true  | string            | 服务名标识                       |
| hostname | true  | string            | 主机名                           |

*返回结果*

```json
*****成功*****
{
    "code":0,
    "message":""
}
****失败****
{
    "code":-400,
    "message":"-400"
}
```

*CURL*
```shell
curl 'http://127.0.0.1:7171/discovery/cancel' -d "zone=sh1&env=test&appid=provider&hostname=myhostname"
```

### 获取实例fetch

*HTTP*

GET http://HOST/discovery/fetch

*请求参数*

| 参数名   | 必选  | 类型              | 说明                             |
| -------- | ----- | ----------------- | -------------------------------- |
| appid    | true  | string            | 服务名标识                       |
| env      | true  | string            | 环境                             |
| zone     | false  | string            | 可用区，不传返回所有zone的                           |
| status | true  | int            | 拉取某状态服务1.接收流量 2.不接收 3.所有状态                           |

*返回结果*

```json
{
    "code": 0,
    "data": {
        "instances": {
            "zone001": [
                {
                    "zone": "zone001",
                    "env": "uat",
                    "appid": "app_id_0",
                    "hostname": "hostname000000",
                    "color": "",
                    "version": "111",
                    "metadata": {
                        "provider": "",
                        "weight": "10"
                    },
                    "addrs": [
                        "http://172.1.1.1:8080",
                        "gorpc://172.1.1.1:8089"
                    ],
                    "status": 1,
                    "reg_timestamp": 1525948301833084700,
                    "up_timestamp": 1525948301833084700,
                    "renew_timestamp": 1525949202959821300,
                    "dirty_timestamp": 1525948301848680000,
                    "latest_timestamp": 1525948301833084700
                }
            ]
        },
        "latest_timestamp": 1525948301833084700
    }
}
```

*CURL*
```shell
curl 'http://127.0.0.1:7171/discovery/fetch?zone=sh1&env=test&appid=provider&status=1'
```

### 批量获取实例fetchs

*HTTP*

GET http://HOST/discovery/fetchs

*请求参数*

| 参数名   | 必选  | 类型              | 说明                             |
| -------- | ----- | ----------------- | -------------------------------- |
| appid    | true  | []string            | 服务名标识                       |
| env      | true  | string            | 环境                             |
| zone     | false  | string            | 可用区，不传返回所有zone的                           |
| status | true  | int            | 拉取某状态服务1.接收流量 2.不接收 3.所有状态                           |

*返回结果*

```json
{
    "code": 0,
    "data": {
        "app_id_0": {
            "instances": {
                "zone001": [
                    {
                        "zone": "zone001",
                        "env": "uat",
                        "appid": "app_id_0",
                        "hostname": "hostname000000",
                        "color": "",
                        "version": "111",
                        "metadata": {
                            "provider": "",
                            "weight": "10"
                        },
                        "addrs": [
                            "http://172.1.1.1:8080",
                            "gorpc://172.1.1.1:8089"
                        ],
                        "status": 1,
                        "reg_timestamp": 1525948301833084700,
                        "up_timestamp": 1525948301833084700,
                        "renew_timestamp": 1525949202959821300,
                        "dirty_timestamp": 1525948301848680000,
                        "latest_timestamp": 1525948301833084700
                    }
                ]
            },
            "latest_timestamp": 1525948301833084700
        },
        "app_id_1": {
            "instances": {
                "zone001": [
                    {
                        "zone": "zone001",
                        "env": "uat",
                        "appid": "app_id_1",
                        "hostname": "hostname111111",
                        "color": "",
                        "version": "222",
                        "metadata": {
                            "provider": "",
                            "weight": "10"
                        },
                        "addrs": [
                            "http://172.1.1.1:7070",
                            "gorpc://172.1.1.1:7079"
                        ],
                        "status": 1,
                        "reg_timestamp": 1525948301833084700,
                        "up_timestamp": 1525948301833084700,
                        "renew_timestamp": 1525949202959821300,
                        "dirty_timestamp": 1525948301848680000,
                        "latest_timestamp": 1525948301833084700
                    }
                ]
            },
            "latest_timestamp": 1525948297987066600
        }
    }
}
```

*CURL*
```shell
curl 'http://127.0.0.1:7171/discovery/fetchs?zone=sh1&env=test&appid=provider&appid=provider2&status=1'
```

### 长轮询获取实例poll

*HTTP*

GET http://HOST/discovery/poll

*请求参数*

| 参数名   | 必选  | 类型              | 说明                             |
| -------- | ----- | ----------------- | -------------------------------- |
| appid    | true  | string            | 服务名标识                       |
| env      | true  | string            | 环境                             |
| zone     | false  | string            | 可用区，不传返回所有zone的                           |
| latest_timestamp | false  | int            | 服务最新更新时间                           |

*返回结果*

```json
{
    "code": 0,
    "data": {
        "instances": {
            "zone001": [
                {
                    "zone": "zone001",
                    "env": "uat",
                    "appid": "app_id_0",
                    "hostname": "hostname000000",
                    "color": "",
                    "version": "111",
                    "metadata": {
                        "provider": "",
                        "weight": "10"
                    },
                    "addrs": [
                        "http://172.1.1.1:8080",
                        "gorpc://172.1.1.1:8089"
                    ],
                    "status": 1,
                    "reg_timestamp": 1525948301833084700,
                    "up_timestamp": 1525948301833084700,
                    "renew_timestamp": 1525949202959821300,
                    "dirty_timestamp": 1525948301848680000,
                    "latest_timestamp": 1525948301833084700
                }
            ]
        },
        "latest_timestamp": 1525948301833084700
    }
}
```

*CURL*
```shell
curl 'http://127.0.0.1:7171/discovery/poll?zone=sh1&env=test&appid=provider&latest_timestamp=0'
```

### 长轮询批量获取实例polls

*HTTP*

GET http://HOST/discovery/polls

*请求参数*

| 参数名   | 必选  | 类型              | 说明                             |
| -------- | ----- | ----------------- | -------------------------------- |
| appid    | true  | []string            | 服务名标识                       |
| env      | true  | string            | 环境                             |
| zone     | false  | string            | 可用区，不传返回所有zone的                           |
| latest_timestamp | false  | []int            | 服务最新更新时间，要与appid一一对应           |

*返回结果*

```json
{
    "code": 0,
    "data": {
        "app_id_0": {
            "instances": {
                "zone001": [
                    {
                        "zone": "zone001",
                        "env": "uat",
                        "appid": "app_id_0",
                        "hostname": "hostname000000",
                        "color": "",
                        "version": "111",
                        "metadata": {
                            "provider": "",
                            "weight": "10"
                        },
                        "addrs": [
                            "http://172.1.1.1:8080",
                            "gorpc://172.1.1.1:8089"
                        ],
                        "status": 1,
                        "reg_timestamp": 1525948301833084700,
                        "up_timestamp": 1525948301833084700,
                        "renew_timestamp": 1525949202959821300,
                        "dirty_timestamp": 1525948301848680000,
                        "latest_timestamp": 1525948301833084700
                    }
                ]
            },
            "latest_timestamp": 1525948301833084700
        },
        "app_id_1": {
            "instances": {
                "zone001": [
                    {
                        "zone": "zone001",
                        "env": "uat",
                        "appid": "app_id_1",
                        "hostname": "hostname111111",
                        "color": "",
                        "version": "222",
                        "metadata": {
                            "provider": "",
                            "weight": "10"
                        },
                        "addrs": [
                            "http://172.1.1.1:7070",
                            "gorpc://172.1.1.1:7079"
                        ],
                        "status": 1,
                        "reg_timestamp": 1525948301833084700,
                        "up_timestamp": 1525948301833084700,
                        "renew_timestamp": 1525949202959821300,
                        "dirty_timestamp": 1525948301848680000,
                        "latest_timestamp": 1525948301833084700
                    }
                ]
            },
            "latest_timestamp": 1525948297987066600
        }
    }
}
```

*CURL*
```shell
curl 'http://127.0.0.1:7171/discovery/polls?zone=sh1&env=test&appid=provider1&appid=provider2&latest_timestamp=01&latest_timestamp=02'
```

### 获取node节点

*HTTP*

GET http://HOST/discovery/nodes

*请求参数*

无

*返回结果*

```json
{
    "code": 0,
    "data": [
        {
            "addr": "172.1.1.1:7171",
            "status": 0,
            "zone": "zone001"
        },
        {
            "addr": "172.1.1.2:7171",
            "status": 0,
            "zone": "zone001"
        },
        {
            "addr": "172.1.1.3:7171",
            "status": 0,
            "zone": "zone001"
        }
    ]
}
```

*CURL*
```shell
curl 'http://127.0.0.1:7171/discovery/nodes'
```

### 修改实例信息set

*HTTP*

POST http://HOST/discovery/set

*请求参数*

| 参数名   | 必选  | 类型              | 说明                             |
| -------- | ----- | ----------------- | -------------------------------- |
| zone     | true  | string            | 可用区                           |
| env      | true  | string            | 环境                             |
| appid    | true  | string            | 服务名标识                       |
| hostname | true  | []string            | 主机名                           |
| status   | false | []int               | 状态，1表示接收流量，2表示不接收 |
| color    | false | []string            | 灰度或集群标识                   |
| metadata | false | []string | 业务自定义信息         string 必须为map[strinng]string 的json格式   |      

*返回结果*

```json
*****成功*****
{
    "code":0,
    "message":""
}
****失败****
{
    "code":-400,
    "message":"-400"
}
```

*CURL*
```shell
curl 'http://127.0.0.1:7171/discovery/set' -d "zone=sh1&env=test&appid=provider&hostname=myhostname&status=1&color=red&hostname=myhostname2&status=1&color=red"
```
