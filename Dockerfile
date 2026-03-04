# Stage 1: Build Tailwind CSS (runs natively, output is arch-independent)
FROM --platform=$BUILDPLATFORM node:22-alpine AS tailwind
WORKDIR /app
COPY package.json ./
RUN npm install
COPY input.css ./
COPY views/ ./views/
RUN npx @tailwindcss/cli -i input.css -o assets/styles.css --minify

# Stage 2: Generate templ + Build Go binary (native compilation with CGO)
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git build-base
RUN go install github.com/a-h/templ/cmd/templ@latest
WORKDIR /app

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN templ generate
COPY --from=tailwind /app/assets/styles.css ./assets/styles.css
RUN go build \
    -ldflags="-s -w -X github.com/rezuscloud/platform-website/version.Version=${VERSION} \
    -X github.com/rezuscloud/platform-website/version.GitCommit=${GIT_COMMIT} \
    -X github.com/rezuscloud/platform-website/version.BuildTime=${BUILD_TIME}" \
    -o /bin/server .

# Stage 3: Production image (requires libc for CGO binary)
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /
COPY --from=builder /bin/server /server
COPY --from=builder /app/assets/ /assets/
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/server"]
