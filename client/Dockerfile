# Build stage
FROM golang:latest AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY client client 
COPY pkg pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o /main /app/client/cmd

# Run stage
FROM alpine:latest

COPY --from=build /main .
COPY client/conf.yaml .

ENTRYPOINT ["/main"]