FROM gliderlabs/alpine:3.2
ENV GOPATH /go
ENV GOBIN $GOPATH/bin

COPY . /go/src/github.com/influxdata/telegraf
RUN apk-install -t build-deps make go git mercurial
RUN cd /go/src/github.com/influxdata/telegraf \
	&& make
RUN rm -rf /go \
	&& apk del --purge build-deps
