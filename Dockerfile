FROM gliderlabs/alpine:3.2

COPY . /go/src/github.com/influxdata/telegraf
RUN apk-install -t build-deps go git mercurial
RUN cd /go/src/github.com/influxdata/telegraf \
	&& export GOPATH=/go \
	&& go get \
	&& go build -ldflags "-s -w" -o /bin/telegraf
RUN rm -rf /go \
	&& apk del --purge build-deps
