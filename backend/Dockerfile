FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /rice-api ./cmd/api

FROM alpine:3.19
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
WORKDIR /app
COPY --from=builder /rice-api .
USER 1000:1000
EXPOSE 8080
CMD ["./rice-api"]
