module github.com/bilibili/discovery

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/bilibili/kratos v0.1.0
	github.com/golang/protobuf v1.3.1 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20190430165422-3e4dfb77656c // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/nbio/st v0.0.0-20140626010706-e9e8d9816f32 // indirect
	github.com/smartystreets/assertions v0.0.0-20190401211740-f487f9de1cd3 // indirect
	github.com/smartystreets/goconvey v0.0.0-20180222194500-ef6db91d284a
	golang.org/x/sync v0.0.0-20181108010431-42b317875d0f
	golang.org/x/sys v0.0.0-20190222072716-a9d3bda3a223 // indirect
	google.golang.org/grpc v1.20.1
	gopkg.in/h2non/gock.v1 v1.0.8
)

replace (
	golang.org/x/net => github.com/golang/net v0.0.0-20190311183353-d8887717615a
	golang.org/x/sync => github.com/golang/sync v0.0.0-20181108010431-42b317875d0f
	golang.org/x/sys => github.com/golang/sys v0.0.0-20180905080454-ebe1bf3edb33
	google.golang.org/grpc => github.com/grpc/grpc-go v1.20.1
)
