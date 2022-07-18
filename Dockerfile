FROM golang:1.17-alpine

RUN apk update && apk add --no-cache git && apk add --no-cach bash && apk add build-base

RUN mkdir -p /app/source
WORKDIR /app/source

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...
RUN go build -o ../build
RUN rm -rf ./*

EXPOSE 8089

CMD [ "../build" ]
