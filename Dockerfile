ARG ARCH="amd64"
ARG OS="linux"
FROM quay.io/prometheus/busybox:latest

COPY etherpad_exporter /bin/etherpad_exporter

ENTRYPOINT ["/bin/etherpad_exporter"]
EXPOSE 9301