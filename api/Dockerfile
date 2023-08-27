# Build stage
FROM golang:1.17 AS build

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o app

# Final stage
FROM alpine:latest

WORKDIR /app
COPY --from=build /app/app .

CMD ["./app"]

