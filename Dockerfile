FROM golang:latest as builder
RUN mkdir /build
RUN mkdir -p /etc/hivehome_exporter
ADD . /build/
WORKDIR /build
RUN go get -d -v ./...
RUN go build ./hivehome_exporter.go


FROM scratch

COPY --from=builder /etc/hivehome_exporter /etc/hivehome_exporter
COPY --from=builder /build/hivehome_exporter /bin/hivehome_exporter
COPY config.yaml  /etc/hivehome_exporter

EXPOSE     8000:8000
ENTRYPOINT [ "/bin/hivehome_exporter" ]