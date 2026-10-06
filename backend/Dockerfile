FROM golang:1.20-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org
COPY . .
RUN go build -o /usr/local/bin/nombresmad ./
EXPOSE 8080
CMD ["/usr/local/bin/nombresmad"]
