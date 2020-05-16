FROM golang:alpine as build

WORKDIR /src/app

USER root

RUN apk update && \
    apk add ffmpeg git

COPY ./main.go ./

RUN go get github.com/jtguibas/cinema

RUN go build

CMD ["./app"]