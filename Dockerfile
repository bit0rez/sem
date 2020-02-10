FROM golang:1.12-alpine as builder

RUN apk update \
    && apk add git gcc libc-dev make

WORKDIR /src
COPY Makefile go.mod go.sum /src/
ARG GOPROXY='direct'
RUN GOPROXY=$GOPROXY go mod download
COPY main.go /src
COPY internal /src/internal
ARG VERSION='latest'
RUN GOPROXY=$GOPROXY GOOS=linux GOARCH=amd64 CGO_ENABLED=1 make binary

FROM scratch

COPY --from=builder /src/sem /sem

EXPOSE 9080
ENTRYPOINT ["/sem"]
