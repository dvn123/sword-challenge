# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go mod download

COPY *.go ./

RUN go build .

EXPOSE 8080

CMD [ "./sword-challenge" ]