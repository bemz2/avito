FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

RUN apk add --no-cache bash

COPY . .

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0

RUN go build -o main cmd/api/main.go

EXPOSE 8080

CMD /bin/bash -c "sleep 5 && ./main" 