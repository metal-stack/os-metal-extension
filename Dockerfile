FROM golang:1.24 AS builder

WORKDIR /work
COPY . .
RUN make build

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /work/os-metal /os-metal
CMD ["/os-metal"]
