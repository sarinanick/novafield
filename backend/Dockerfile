FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o novafield-api .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/novafield-api .
RUN mkdir -p uploads && chown -R appuser:appgroup /app
USER appuser
EXPOSE 3001
CMD ["./novafield-api"]
