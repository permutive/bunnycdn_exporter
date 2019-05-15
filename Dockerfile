FROM        quay.io/prometheus/busybox:latest
LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

COPY bunnycdn_exporter /bin/bunnycdn_exporter

ENTRYPOINT ["/bin/bunnycdn_exporter"]
EXPOSE     9584
