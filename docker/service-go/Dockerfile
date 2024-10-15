FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./main ./cmd/server/main.go

EXPOSE 6060

CMD ["./wait-for-it.sh", "rabbit:15672", "--"]
