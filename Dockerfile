# --- Tailwind CSS ---
FROM node:20-alpine AS assets
WORKDIR /src/frontend
COPY frontend/package.json frontend/tailwind.config.js ./
COPY frontend/templates ./templates
COPY frontend/static/css/input.css ./static/css/input.css
RUN npm install \
  && npx tailwindcss -i ./static/css/input.css -o ./static/css/output.css --minify

# --- Go binary ---
FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=assets /src/frontend/static/css/output.css ./frontend/static/css/output.css
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/findus ./backend/app

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
