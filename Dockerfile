#############      builder-base                             #############
FROM golang:1.13 AS builder
WORKDIR /work
COPY . .
# RUN go mod download
RUN hack/install-requirements.sh
RUN make VERIFY=$VERIFY all

#############      base                                     #############
FROM alpine:3.11 AS base
WORKDIR /
COPY charts /charts

COPY --from=builder /work/os-metal /os-metal
CMD ["/os-metal"]
