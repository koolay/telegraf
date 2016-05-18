FROM gliderlabs/alpine:3.2

COPY . /go/src/github.com/influxdata/telegraf
RUN apk-install -t build-deps go git mercurial \
	&& cd /go/src/github.com/influxdata/telegraf \
	&& export GOPATH=/go \
	&& go get \
	&& go build -ldflags "-X main.Version $(cat VERSION)" -o /bin/telegraf \
	&& rm -rf /go \
	&& apk del --purge build-deps
