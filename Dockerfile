FROM golang:alpine

COPY ./src/ /go/src/
COPY ./build.sh /go/
COPY ./docker-entrypoint.sh /

RUN apk add --update bash && apk add --update curl && apk add --update git &&\
    rm -rf /var/cache/apk/*
RUN curl https://glide.sh/get | sh

RUN ./build.sh

ENTRYPOINT ["/docker-entrypoint.sh"]