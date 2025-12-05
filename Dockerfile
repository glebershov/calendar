FROM golang:latest AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o calendar ./cmd/calendar

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/calendar /app/calendar
COPY config.yaml /app/config.yaml
COPY --from=builder /app/migrations /app/migrations
EXPOSE 8080
CMD ["/app/calendar"]