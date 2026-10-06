FROM golang:1.27.0-alpine AS builder

WORKDIR /audioaDialog

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o audioaDialog ./cmd/audioaDialog

FROM alpine:latest

WORKDIR /files

COPY --from=builder /files/audioaDialog .

EXPOSE 8086

CMD ["./audioaDialog"]