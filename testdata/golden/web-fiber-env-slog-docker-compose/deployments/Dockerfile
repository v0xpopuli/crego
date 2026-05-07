FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/orders-api ./cmd/orders-api

FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app && mkdir -p /app/data && chown -R app:app /app

WORKDIR /app

COPY --from=build /out/orders-api /app/orders-api
USER app

EXPOSE 8080

ENTRYPOINT ["/app/orders-api"]
