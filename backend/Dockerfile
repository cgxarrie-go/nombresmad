FROM golang:1.20-alpine
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN go build -mod=vendor -o /usr/local/bin/nombresmad ./
EXPOSE 8080
CMD ["/usr/local/bin/nombresmad"]
