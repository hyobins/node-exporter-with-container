ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox-${OS}-${ARCH}:glibc
#FROM alpine:3.12.4
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

ARG ARCH="amd64"
ARG OS="linux"
#COPY .build/${OS}-${ARCH}/node_exporter /bin/node_exporter
COPY node_exporter /bin/node_exporter

EXPOSE      9100
USER        root 
ENTRYPOINT  [ "/bin/node_exporter" ]
