FROM golang:1.22.3-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o ./cmd/url-shortener/main.go


FROM alpine:latest AS runner
WORKDIR /app
COPY --from=builder /cmd/app/url-shortener .
EXPOSE 8080
ENTRYPOINT ["./example-golang"]