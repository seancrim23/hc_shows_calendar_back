# build stage
FROM golang as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# final stage
FROM scratch
COPY --from=builder /app/hc_shows_calendar_back /app/
RUN apt install -y ca-certificates
EXPOSE 8080
ENTRYPOINT ["/app/hc_shows_calendar_back"]