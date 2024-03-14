FROM golang:latest

COPY ./ /app

WORKDIR /app/main_service

RUN go build main.go

ENTRYPOINT ./main ../.env