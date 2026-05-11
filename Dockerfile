# --- Tailwind CSS ---
FROM node:20-alpine AS assets
WORKDIR /src
COPY package.json tailwind.config.js ./
COPY web/templates ./web/templates
COPY web/static/css/input.css ./web/static/css/input.css
RUN npm install \
  && npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify

# --- Go binary ---
FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=assets /src/web/static/css/output.css ./web/static/css/output.css
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/findus ./cmd/findus

# --- Runtime ---
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=build /out/findus /findus
ENV FINDUS_DATA_DIR=/data \
    FINDUS_PORT=8080
EXPOSE 8080
VOLUME ["/data"]
USER nonroot:nonroot
ENTRYPOINT ["/findus"]
