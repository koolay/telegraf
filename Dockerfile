FROM gliderlabs/alpine:3.2

COPY . /go/src/github.com/influxdata/telegraf
RUN apk-install -t build-deps make go git mercurial
RUN cd /go/src/github.com/influxdata/telegraf \
	&& export GOPATH=/go \
	&& make
RUN rm -rf /go \
	&& apk del --purge build-deps
