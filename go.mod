module github.com/toppr-systems/dops

go 1.16

require (
	github.com/alecthomas/kingpin v2.2.5+incompatible
	github.com/aws/aws-sdk-go v1.44.215
	github.com/cloudflare/cloudflare-go v0.13.2
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/linki/instrumented_http v0.3.0
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace k8s.io/klog/v2 => github.com/Raffo/knolog v0.0.0-20211016155154-e4d5e0cc970a
