module github.com/toppr-systems/dops

go 1.16

require (
	github.com/alecthomas/assert v0.0.0-20170929043011-405dbfeb8e38 // indirect
	github.com/alecthomas/colour v0.1.0 // indirect
	github.com/alecthomas/kingpin v2.2.5+incompatible
	github.com/alecthomas/repr v0.0.0-20200325044227-4184120f674c // indirect
	github.com/aws/aws-sdk-go v1.40.53
	github.com/cloudflare/cloudflare-go v0.51.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8
	github.com/linki/instrumented_http v0.3.0
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace k8s.io/klog/v2 => github.com/Raffo/knolog v0.0.0-20211016155154-e4d5e0cc970a
