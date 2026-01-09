FROM golang:1.25-bookworm AS builder

WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GODEBUG=netdns=go

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o advoid main.go


FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=builder /app/advoid /advoid

COPY oisd_big_abp.txt /

ENTRYPOINT ["/advoid"]
