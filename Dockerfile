FROM golang:1.26 AS builder

WORKDIR /work
COPY . .
RUN make build

FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /
COPY --from=builder /work/os-metal /os-metal
CMD ["/os-metal"]
