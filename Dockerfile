FROM alpine:3.6
LABEL maintainer="Jose Armesto <jose@armesto.net>"

EXPOSE 8000

ENTRYPOINT ["/opt/app"]

RUN apk add --update --repository https://dl-cdn.alpinelinux.org/alpine/edge/community/ tini=0.15.0-r0 ca-certificates && \
    addgroup -S app && adduser -u 10001 -S -g /opt/app app && \
    rm -rf /var/cache/apk/* /tmp/*

USER app
ADD app /opt/app
