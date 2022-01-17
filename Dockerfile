FROM  quay.io/prometheus/busybox:latest
LABEL maintainer="Ricardo SRE <sre@ricardo.ch>"

COPY --chown=1984:1984 bunnycdn_exporter /bin/bunnycdn_exporter

USER 1984
ENTRYPOINT ["/bin/bunnycdn_exporter"]
EXPOSE     9584
