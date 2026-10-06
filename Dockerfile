FROM golang:1.27.0-alpine AS builder

WORKDIR /audioDialog

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o audioDialog ./cmd/audioDialog

FROM alpine:latest

WORKDIR /audioDialog

COPY --from=builder /audioDialog/audioDialog .

EXPOSE 8086

CMD ["./audioDialog"]