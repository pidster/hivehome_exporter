FROM busybox

RUN mkdir -p /etc/hivehome_exporter

COPY hivehome_exporter  /bin/hivehome_exporter
COPY config.yaml  /etc/hivehome_exporter

EXPOSE     8000:8000
ENTRYPOINT [ "/bin/hivehome_exporter" ]