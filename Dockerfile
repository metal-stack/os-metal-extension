#############      builder-base                             #############
FROM golang:1.12.6 AS builder
WORKDIR /work
COPY . .
RUN go mod download

RUN make VERIFY=$VERIFY all

#############      base                                     #############
FROM alpine:3.9 AS base
WORKDIR /
COPY charts /charts

COPY --from=builder /work/os-metal /os-metal
CMD ["/os-metal"]
