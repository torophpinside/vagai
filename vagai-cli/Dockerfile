FROM golang:1.21-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o vagai .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /build/vagai .
COPY configs/ configs/
COPY curriculos/ curriculos/

RUN chmod +x vagai

ENTRYPOINT ["./vagai"]
