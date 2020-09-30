FROM golang:1.15-alpine AS builder
RUN mkdir -p /k8s-event-listener
WORKDIR /k8s-event-listener

RUN apk add -u git curl

COPY go.* ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/k8s-event-listener

FROM alpine

LABEL maintainer="Werkspot <technology@werkspot.com>"

COPY --from=builder /k8s-event-listener/bin/. .

CMD ["/k8s-event-listener"]