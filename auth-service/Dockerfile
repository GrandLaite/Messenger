FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service cmd/main.go

FROM alpine:3.19
WORKDIR /srv
COPY --from=build /auth-service /usr/local/bin/auth-service
EXPOSE 8081
ENTRYPOINT ["auth-service"]
