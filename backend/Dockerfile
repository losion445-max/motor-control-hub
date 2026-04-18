FROM golang:1.26.1-alpine AS builder

WORKDIR /user/src/app

COPY go.mod go.sum* ./

RUN go mod download

COPY . .

RUN go build -o application ./cmd/

FROM alpine:3.22

RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup --disabled-password appuser

WORKDIR /app

COPY --from=builder /user/src/app/application .

RUN chmod +x application && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./application"]
