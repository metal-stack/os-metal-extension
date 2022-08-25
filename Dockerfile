FROM golang:1.19-alpine3.16 AS builder
RUN apk add make git gcc musl-dev
WORKDIR /work
COPY . .
RUN make all

FROM alpine:3.16 AS base
WORKDIR /
COPY charts /charts

COPY --from=builder /work/os-metal /os-metal
CMD ["/os-metal"]
