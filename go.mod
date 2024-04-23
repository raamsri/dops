module github.com/toppr-systems/dops

go 1.16

require (
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38 // indirect
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/kingpin v2.2.5+incompatible
	github.com/alecthomas/repr v0.0.0-20200325044227-4184120f674c // indirect
	github.com/aws/aws-sdk-go v1.40.53
	github.com/cloudflare/cloudflare-go v0.13.2
	github.com/google/go-cmp v0.6.0
	github.com/linki/instrumented_http v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.19.0
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
)

replace k8s.io/klog/v2 => github.com/Raffo/knolog v0.0.0-20211016155154-e4d5e0cc970a
