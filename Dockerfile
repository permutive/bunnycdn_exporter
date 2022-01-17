FROM  quay.io/prometheus/busybox:latest
LABEL maintainer="Ricardo SRE <sre@ricardo.ch>"

USER 1984

COPY --chown=1984:1984 bunnycdn_exporter /bin/bunnycdn_exporter

ENTRYPOINT ["/bin/bunnycdn_exporter"]
EXPOSE     9584
