FROM golang:1.21

WORKDIR /app

COPY /app/go.mod /app/go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/app/cmd/web/main /app/app/cmd/web/main.go

EXPOSE 8080

CMD ["/app/app/cmd/web/main"]
