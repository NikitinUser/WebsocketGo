FROM golang:1.21

WORKDIR /app

COPY . .

RUN go build -o ./local_user_service ./app/cmd/web/local_user_service/main.go

EXPOSE 6070

CMD ["./local_user_service"]
