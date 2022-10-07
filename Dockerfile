FROM golang:1.18-alpine as build_base
WORKDIR /src
COPY go.mod .
COPY go.sum .

RUN apk add --no-cache gcc libc-dev
RUN go mod download

FROM build_base AS builder

ADD . /src
WORKDIR /src
RUN go build --tags musl -o /src/bin/ ./cmd/...

FROM alpine:latest

WORKDIR /app

RUN mkdir /app/config

COPY --from=builder /src/bin /app