FROM golang:1.23-alpine as builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o kanban-board ./cmd/main.go

FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/kanban-board .

RUN chmod +x /root/kanban-board

CMD ["./kanban-board"]