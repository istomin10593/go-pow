# Build stage
FROM golang:latest AS build

WORKDIR /app

# COPY go.mod go.sum ./

# COPY cmd/*.go .
COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

# Run stage
FROM alpine:latest

COPY --from=build /main .
COPY conf.yaml .

ENTRYPOINT ["/main"]