FROM golang:1.23-alpine3.21 AS builder
RUN apk add make git gcc musl-dev
WORKDIR /work
COPY . .
RUN make build

FROM alpine:3.21 AS base
WORKDIR /
COPY charts /charts

COPY --from=builder /work/os-metal /os-metal
CMD ["/os-metal"]
