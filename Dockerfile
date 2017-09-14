FROM golang:1.8
ADD workspace /go
RUN go get github.com/golang/lint/golint
RUN make lint test build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/bin/linux/server/k8s-aws-cloudwatchlogs /usr/local/bin/k8s-aws-cloudwatchlogs
CMD ["k8s-aws-cloudwatchlogs"]
