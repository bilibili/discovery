# Discovery [![Build Status](https://travis-ci.org/Bilibili/discovery.svg?branch=master)](https://travis-ci.org/Bilibili/discovery)

Discovery is a based service that is production-ready and primarily used at [Bilibili](https://www.bilibili.com/) for locating services for the purpose of load balancing and failover of middle-tier servers.

## 快速开始

### 环境

请使用go1.9.x及以上版本

### 构建
```shell
cd $GOPATH
mkdir -p github.com/Bilibili
cd $GOPATH/github.com/Bilibili
git clone https://github.com/Bilibili/discovery.git
cd discovery/cmd/discovery
go build
```

### 运行
```shell
./discovry -conf discovery-example.toml
```

### 配置文件解读

请详细查看`discovery-example.toml`内注释说明

### 客户端 

* [API文档](api.md)
* [GoSDK](naming/client.go)
* [JavaSDK](https://github.com/flygit/discoveryJavaSDK)

## 背景及设计

[点我了解背后的故事](intro.md)

## 反馈及讨论

请加微信

<img width="200" height="200" src="discovery_wechat.png" />
