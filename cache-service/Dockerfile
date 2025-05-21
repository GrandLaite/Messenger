FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cache-service ./cmd/main.go

FROM alpine:3.19
WORKDIR /srv
COPY --from=build /cache-service /usr/local/bin/cache-service
EXPOSE 8085
ENTRYPOINT ["cache-service"]
