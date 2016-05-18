FROM gliderlabs/alpine:3.2

COPY . /go/src/github.com/influxdata/telegraf
RUN apk-install -t build-deps go git mercurial \
	&& cd /go/src/github.com/influxdata/telegraf \
	&& export GOPATH=/go

RUN	go get \
	&& go build -ldflags "-s -w" -o /bin/telegraf

RUN rm -rf /go \
	&& apk del --purge build-deps
