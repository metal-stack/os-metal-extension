FROM golang:1.18 AS builder
WORKDIR /work
COPY . .
RUN make

FROM alpine:3.16 AS base
WORKDIR /
COPY charts /charts

COPY --from=builder /work/os-metal /os-metal
CMD ["/os-metal"]
