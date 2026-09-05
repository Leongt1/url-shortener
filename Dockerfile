# --- Stage 1: build ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

# copy dependencies and install them
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# --- Stage 2: runtime ---
FROM alpine:3.22

COPY --from=builder /app/server /server

EXPOSE 8080

CMD [ "/server" ]