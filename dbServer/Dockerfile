FROM golang:1.23 AS deploy-builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -trimpath -ldflags "-w -s" -o app

#------------------------------------------------------------
FROM debian:bookworm-slim AS deploy

RUN apt-get update&& \
apt-get install -y ca-certificates && \
update-ca-certificates

COPY --from=deploy-builder /app/app .

CMD ["./app"]

#------------------------------------------------------------
FROM golang:1.23 AS dev

WORKDIR /app

RUN go install github.com/gin-gonic/gin

CMD ["./app"]