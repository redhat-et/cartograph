FROM golang:1.25 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cartograph ./cmd/cartograph/

FROM registry.access.redhat.com/ubi9-minimal:latest
COPY --from=builder /cartograph /usr/local/bin/cartograph
USER 1001
ENTRYPOINT ["cartograph"]
