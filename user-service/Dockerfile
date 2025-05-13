FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /user-service cmd/main.go

FROM alpine:3.19
WORKDIR /srv
COPY --from=build /user-service /usr/local/bin/user-service
EXPOSE 8082
ENTRYPOINT ["user-service"]
