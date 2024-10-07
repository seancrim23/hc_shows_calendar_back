# build stage
FROM golang:1.20 as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
RUN apt-get update && apt-get install ca-certificates -y
# final stage
FROM scratch
COPY --from=builder /app/hc_shows_calendar_back /app/
EXPOSE 8080
ENTRYPOINT ["/app/hc_shows_calendar_back"]