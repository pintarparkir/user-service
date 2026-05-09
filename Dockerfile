# Multi-stage Dockerfile — one image per service via SERVICE build-arg
# Build: docker build --build-arg SERVICE=reservation -t parkirpintar/reservation .
ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-alpine AS builder
ARG SERVICE
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/app ./cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/app /app
USER nonroot:nonroot
EXPOSE 8080 9090
ENTRYPOINT ["/app"]
