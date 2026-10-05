# Stage 1: Build the UI
FROM node:22-alpine AS ui-builder

# Shared widget builder package (consumed by the UI via a file: dependency)
WORKDIR /app/widget-builder
COPY widget-builder/package.json widget-builder/package-lock.json ./
RUN npm ci
COPY widget-builder/ ./
RUN npm run build

WORKDIR /app/ui
COPY ui/package.json ui/package-lock.json ui/.npmrc ./
RUN npm ci

# Copy OpenAPI spec for TS client generation
COPY spec/tsp-output/schema/openapi.yaml /app/spec/tsp-output/schema/openapi.yaml
COPY ui/ ./

RUN npm run generate:api && npm run build

# Stage 2: Build the Go server
FROM golang:1.26-alpine AS server-builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/
COPY tools.go ./

RUN CGO_ENABLED=1 go build -o /app/server ./cmd/server/...

# Stage 3: Minimal runtime
FROM alpine:3.21

RUN apk add --no-cache ca-certificates
RUN mkdir -p /data

WORKDIR /app
COPY --from=server-builder /app/server ./server
COPY --from=ui-builder /app/ui/dist/ ./ui/dist/

VOLUME ["/data"]
EXPOSE 8080

ENTRYPOINT ["/app/server"]
