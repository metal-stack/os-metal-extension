FROM golang:1.22-alpine3.20 AS builder
RUN apk add make git gcc musl-dev
WORKDIR /work
COPY . .
RUN make all

FROM alpine:3.20 AS base
WORKDIR /
COPY charts /charts

COPY --from=builder /work/os-metal /os-metal
CMD ["/os-metal"]
