FROM golang:latest

COPY ./ /app

WORKDIR /app/stat_service
RUN go mod download -x
RUN go build main.go

ENTRYPOINT ./main