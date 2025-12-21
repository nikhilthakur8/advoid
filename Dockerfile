FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o advoid main.go


FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=builder /app/advoid /advoid

COPY oisd_big_abp.txt /

ENTRYPOINT ["/advoid"]
