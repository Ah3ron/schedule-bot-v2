FROM golang:1.23.6-alpine3.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o schedule-bot ./main.go

FROM scratch

COPY --from=builder /app/schedule-bot /schedule-bot

ENTRYPOINT ["/schedule-bot"]