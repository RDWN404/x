# Gunakan image resmi Golang untuk membangun aplikasi
FROM golang:1.23.3 AS builder

# Set working directory di dalam container
WORKDIR /app

# Salin go.mod dan go.sum lalu instal dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Salin seluruh kode aplikasi ke dalam container
COPY . .

# Build aplikasi Golang
RUN go build -o x .

# Gunakan image ringan untuk menjalankan aplikasi
FROM debian:bullseye-slim

# Set working directory di dalam container
WORKDIR /root/

# Salin file aplikasi dari stage builder
COPY --from=builder /app/x .

# Expose port 8080 untuk aplikasi
EXPOSE 8081

# Jalankan aplikasi
CMD ["./x"]
