FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /message-service cmd/main.go

FROM alpine:3.19
WORKDIR /srv
COPY --from=build /message-service /usr/local/bin/message-service
EXPOSE 8083
ENTRYPOINT ["message-service"]
