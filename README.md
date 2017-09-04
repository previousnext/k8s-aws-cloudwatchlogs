Kubernetes: AWS CloudWatch Logs
===============================

[![CircleCI](https://circleci.com/gh/previousnext/k8s-pagerduty.svg?style=svg)](https://circleci.com/gh/previousnext/k8s-pagerduty)

![Diagram](/docs/diagram.png "Diagram")

DaemonSet for pushing logs to AWS CloudWatch Logs.

* Load logs from `/var/log/containers`
* Push to AWS CloudWatchLogs
 * New files will push all lines
 * Existing files will only push new lines
* Leverages Docker inbuilt AWS CloudWatch Logs client for handling CloudWatch limits
  * https://github.com/moby/moby/blob/master/daemon/logger/awslogs
  * http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_limits.html

## Usage

```bash
kubectl create -f kubernetes/daemonset.yaml
```

## Development

```bash
# Run the test suite
cd workspace && make lint
cd workspace && make test

# Build the project
cd workspace && make build

# Release the project
make release VERSION=1.0.0
```
