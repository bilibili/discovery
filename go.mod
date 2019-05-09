module discovery

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/bilibili/discovery v1.0.1
	github.com/gin-gonic/gin v0.0.0-20180512030042-bf7803815b0b
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/smartystreets/goconvey v0.0.0-20180222194500-ef6db91d284a
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c // indirect
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20190508220229-2d0786266e9c // indirect
	google.golang.org/grpc v1.20.1
	gopkg.in/h2non/gock.v1 v1.0.8
)

replace (
	golang.org/x/net => github.com/golang/net v0.0.0-20190311183353-d8887717615a
	golang.org/x/sync => github.com/golang/sync v0.0.0-20181108010431-42b317875d0f
	golang.org/x/sys => github.com/golang/sys v0.0.0-20180905080454-ebe1bf3edb33
	google.golang.org/grpc => github.com/grpc/grpc-go v1.20.1
)
