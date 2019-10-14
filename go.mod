module github.com/freeconf/bridge

go 1.12

require (
	github.com/freeconf/manage v0.0.0-20190928152552-c94a450e817a
	github.com/freeconf/yang v0.0.0-20190915134354-9d96c3c868e8
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/nlopes/slack v0.1.0
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
)

replace github.com/freeconf/yang => ../yang

replace github.com/freeconf/manage => ../manage
