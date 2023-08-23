#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go build -o /go/bin/app -v ./...

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/app /app
ENV NAMESPACE default
RUN echo "namespace: $NAMESPACE"
ENTRYPOINT /app -ns=$NAMESPACE
LABEL Name=k8s-fix-unhealthy-pod Version=0.0.1
LABEL org.opencontainers.image.source="https://github.com/heavenwing/k8s-fix-unhealthy-pod"