FROM alpine:latest

ADD go-app /go-app
ADD dump.csv /dump.csv

ENTRYPOINT ["/go-app"]
