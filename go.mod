module github.com/kubeless/kafka-trigger

go 1.12

require (
	github.com/Shopify/sarama v1.27.0
	github.com/cenkalti/backoff/v4 v4.0.2
	github.com/coreos/prometheus-operator v0.0.0-20171201110357-197eb012d973
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/imdario/mergo v0.3.7
	github.com/kubeless/kubeless v1.0.7
	github.com/mitchellh/gox v1.0.1 // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v0.0.3
	golang.org/x/sys v0.0.0-20200519105757-fe76b779f299 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/api v0.0.0-20180308224125-73d903622b73
	k8s.io/apiextensions-apiserver v0.0.0-20180327033742-750feebe2038
	k8s.io/apimachinery v0.0.0-20180228050457-302974c03f7e
	k8s.io/client-go v7.0.0+incompatible
)
