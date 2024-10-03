FROM golang:1.22.2

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY /cmd/webserver/*.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /application

CMD ["/application"]