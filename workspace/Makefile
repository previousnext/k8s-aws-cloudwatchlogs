#!/usr/bin/make -f

lint:
	golint -set_exit_status src/github.com/previousnext/k8s-aws-cloudwatchlogs/...

test:
	./hack/test.sh github.com/previousnext/k8s-aws-cloudwatchlogs/...

build:
	./hack/build.sh linux server k8s-aws-cloudwatchlogs github.com/previousnext/k8s-aws-cloudwatchlogs
