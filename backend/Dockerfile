FROM golang:1.24-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@v1.61.7

COPY . .

EXPOSE 9090

CMD ["air", "-c", ".air.toml"]
