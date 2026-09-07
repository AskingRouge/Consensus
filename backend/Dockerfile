# Multi-stage build: compile a static Go binary, then ship it in a minimal
# distroless-style image. Works locally (`docker build .`) and is what
# Azure Container Apps will pull from a registry once deployment (Phase 8)
# is wired up.

FROM golang:1.22-alpine AS build
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/consensus-backend .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/consensus-backend .

EXPOSE 8080
ENTRYPOINT ["./consensus-backend"]
