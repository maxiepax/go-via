FROM golang:1.16

ADD go-via /usr/local/bin/go-via

EXPOSE 8080

ENTRYPOINT ["go-via"]
