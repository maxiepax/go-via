FROM amd64/alpine:latest

RUN apk add --no-cache bash libc6-compat
ADD go-via /usr/local/bin/go-via

EXPOSE 8080

ENTRYPOINT ["go-via"]
