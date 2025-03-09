FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o main ./cmd/server

FROM alpine
WORKDIR /app
COPY --from=build /app/main .
COPY .env .
EXPOSE 3000
CMD ["./main"]
