FROM golang as builder
COPY . /go/src/github.com/b1-systems/bareos_exporter
WORKDIR /go/src/github.com/b1-systems/bareos_exporter
RUN go get -v .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bareos_exporter .

FROM busybox:latest

WORKDIR /bareos_exporter
COPY --from=builder /go/src/github.com/b1-systems/bareos_exporter/bareos_exporter bareos_exporter

ENTRYPOINT ["./bareos_exporter"]
EXPOSE $port
