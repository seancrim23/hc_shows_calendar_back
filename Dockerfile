# build stage
FROM golang AS builder

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

# set local variables to connect to firestore
ENV GCP_PROJECT_ID="hc-show-calendar"
ENV FIRESTORE_EMULATOR_HOST="localhost:5050"

COPY --from=builder /app/hc_shows_calendar_back /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
ENTRYPOINT ["/app/hc_shows_calendar_back"]