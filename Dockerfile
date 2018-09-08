FROM alpine:latest

RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/*
RUN update-ca-certificates

ADD go-app /go-app

ENTRYPOINT ["/go-app"]
