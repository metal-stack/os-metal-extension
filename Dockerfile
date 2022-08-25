FROM golang:1.19 AS builder
WORKDIR /work
COPY . .
RUN make all

FROM alpine:3.16 AS base
WORKDIR /
COPY charts /charts

COPY --from=builder /work/os-metal /os-metal
CMD ["/os-metal"]
