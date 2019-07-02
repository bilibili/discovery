module github.com/bilibili/discovery

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/gin-contrib/sse v0.0.0-20190301062529-5545eab6dad3 // indirect
	github.com/gin-gonic/gin v0.0.0-20180512030042-bf7803815b0b
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/mattn/go-isatty v0.0.7 // indirect
	github.com/smartystreets/assertions v0.0.0-20190401211740-f487f9de1cd3 // indirect
	github.com/smartystreets/goconvey v0.0.0-20180222194500-ef6db91d284a
	github.com/ugorji/go v1.1.4 // indirect
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	google.golang.org/grpc v1.20.1
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/h2non/gock.v1 v1.0.8
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace (
	golang.org/x/net => github.com/golang/net v0.0.0-20190311183353-d8887717615a
	golang.org/x/sync => github.com/golang/sync v0.0.0-20181108010431-42b317875d0f
	golang.org/x/sys => github.com/golang/sys v0.0.0-20180905080454-ebe1bf3edb33
	google.golang.org/grpc => github.com/grpc/grpc-go v1.20.1
)
