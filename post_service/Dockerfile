FROM golang:latest

COPY ./ /app



WORKDIR /app/post_service

RUN go build main.go

ENTRYPOINT ./main