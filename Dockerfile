# ghcr/production build: see the Makefile's `docker`/`docker-prod` targets.
# `Dockerfile.local` is the near-identical copy
# broker-portal/deploy/docker-compose.yml builds from.
FROM golang:1.22-alpine AS build
ARG VERSION=0.0.0
ARG BUILD_SHA=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.ServiceVersion=${VERSION} -X main.Build=${BUILD_SHA}" -o /out/email-service ./cmd/main.go

FROM alpine:3
WORKDIR /app
COPY --from=build /out/email-service /app/email-service
VOLUME ["/templates"]
CMD ["/app/email-service"]
