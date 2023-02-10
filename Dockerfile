FROM golang:1.20-alpine3.17 AS builder
RUN apk add make git gcc musl-dev
WORKDIR /work
COPY . .
RUN make all

FROM alpine:3.17 AS base
WORKDIR /
COPY charts /charts

COPY --from=builder /work/os-metal /os-metal
CMD ["/os-metal"]
