FROM golang:1.21

WORKDIR /app

COPY /app/go.mod /app/go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./main ./app/cmd/web/main.go

EXPOSE 6060

CMD ["./wait-for-it.sh", "rabbit:15672", "--"]
