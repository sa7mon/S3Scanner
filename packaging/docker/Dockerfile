FROM golang:1.22-alpine AS builder

ARG VERSION=dev

WORKDIR /app
COPY .. /app

RUN go build -ldflags="-X 'main.version=${VERSION}'" -o /s3scanner .
RUN go build -o /mqingest ./cmd/mqingest/

FROM alpine

COPY --from=builder /s3scanner /
COPY --from=builder /mqingest /
ENV PATH="/"
ENTRYPOINT ["/s3scanner"]

