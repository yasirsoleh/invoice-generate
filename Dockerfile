FROM docker.io/golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . /app/
RUN go build -o invoice-generate ./main.go

FROM docker.io/alpine:3.14
WORKDIR /app
RUN apk add tzdata
COPY --from=builder /app/invoice-generate /app
ENTRYPOINT ["/app/invoice-generate"]
