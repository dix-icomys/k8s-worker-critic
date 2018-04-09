FROM alpine:latest

RUN apk add --update ca-certificates

COPY build/critic /usr/local/bin/critic

CMD ["critic"]
