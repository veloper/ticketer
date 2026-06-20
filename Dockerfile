FROM golang:1.25-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o ticketer ./cmd/ticketer && \
    CGO_ENABLED=0 go build -o tktrctl ./cmd/tktrctl

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /data
COPY --from=builder /build/ticketer /usr/local/bin/ticketer
COPY --from=builder /build/tktrctl /usr/local/bin/tktrctl
ENV TICKETER_DB_PATH=/data/ticketer.db
EXPOSE 8300
CMD ["ticketer"]
