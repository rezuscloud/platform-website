# Stage 1: Build Tailwind CSS
FROM node:22-alpine AS tailwind
WORKDIR /app
COPY package.json ./
RUN npm install
COPY input.css ./
COPY views/ ./views/
RUN npx @tailwindcss/cli -i input.css -o assets/styles.css --minify

# Stage 2: Generate templ + Build Go binary
FROM golang:1.24-alpine AS builder
RUN go install github.com/a-h/templ/cmd/templ@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN templ generate
COPY --from=tailwind /app/assets/styles.css ./assets/styles.css
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server .

# Stage 3: Production image
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /bin/server /server
COPY --from=builder /app/assets/ /assets/
EXPOSE 3000
USER nonroot:nonroot
ENTRYPOINT ["/server"]
