# Build stage
FROM golang:latest AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# COPY server/cmd server/cmd
COPY server server
COPY pkg pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /main /app/server/cmd

# Run stage
FROM alpine:latest

COPY --from=build /main .
COPY server/source/quotes.txt source/quotes.txt
COPY server/conf.yaml .

ENTRYPOINT ["/main"]