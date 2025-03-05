FROM golang:1.20 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/agent
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/agent main.go

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/bin/agent /usr/local/bin/agent

ENV COMPUTING_POWER=4
ENV ORCHESTRATOR_URL=http://orchestrator:8080

CMD ["agent"]
