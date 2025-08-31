# ---- Build stage ----
FROM golang:1.24.4-alpine AS builder
WORKDIR /app


# Copy source code
COPY . .

# Build optimized binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server .



# ---- Run stage ----
FROM scratch
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
