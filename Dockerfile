FROM golang:latest as builder

WORKDIR /build
COPY hivehome_exporter.go ./
RUN go get ./...
RUN GOOS=linux go build .


FROM busybox

RUN mkdir -p /etc/hivehome_exporter

COPY --from=builder /build/hivehome_exporter /bin/hivehome_exporter
COPY config.yaml  /etc/hivehome_exporter

EXPOSE     8000:8000
ENTRYPOINT [ "/bin/hivehome_exporter" ]