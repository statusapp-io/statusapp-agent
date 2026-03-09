FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /statusapp-agent .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /statusapp-agent /usr/local/bin/statusapp-agent

ENTRYPOINT ["statusapp-agent"]
