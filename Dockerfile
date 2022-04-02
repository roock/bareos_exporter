FROM golang:1.18.0 as builder

RUN apt-get update \
 && apt-get install -y upx

WORKDIR /go/src/github.com/vierbergenlars/bareos_exporter
COPY . .

RUN go get -v .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bareos_exporter .
RUN upx bareos_exporter

FROM busybox:1.34.1
WORKDIR /bareos_exporter
COPY --from=builder /go/src/github.com/vierbergenlars/bareos_exporter/bareos_exporter bareos_exporter
COPY entrypoint.sh /entrypoint.sh
RUN chmod u+x /entrypoint.sh

ENV PORT 9625
ENV DB_TYPE postgres
ENV DB_HOST localhost
ENV DB_PORT 5432
ENV DB_NAME bareos
ENV DB_USER bareos
ENV SSL_MODE disable
ENV WAIT_FOR_DB 0

EXPOSE $PORT
ENTRYPOINT ["/entrypoint.sh"]
