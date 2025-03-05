FROM golang:1.20 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/orchestrator
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/orchestrator main.go

FROM alpine:3.17
WORKDIR /app

COPY --from=builder /app/bin/orchestrator /usr/local/bin/orchestrator

EXPOSE 8080

ENV PORT=8080

CMD ["orchestrator"]
