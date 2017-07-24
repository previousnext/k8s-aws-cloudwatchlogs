Kubernetes: AWS CloudWatch Logs
===============================

DaemonSet for pushing logs to AWS CloudWatch Logs.

* Uses "labels" to determine CloudWatch Logs "group" and "stream" names.
* Leverages Docker inbuilt AWS CloudWatch Logs client for handling CloudWatch limits
** https://github.com/moby/moby/blob/master/daemon/logger/awslogs
** http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_limits.html

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
