FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /notification-service cmd/main.go

FROM alpine:3.19
WORKDIR /srv
RUN apk add --no-cache ca-certificates
COPY --from=build /notification-service /usr/local/bin/notification-service
EXPOSE 8090
ENTRYPOINT ["notification-service"]
