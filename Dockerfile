FROM golang:1.18.0-bullseye as builder

COPY ./ /tmp/build

RUN cd /tmp/build && CGO_ENABLED=0 go build ./cmd/server

FROM alpine:3.14.0

COPY --from=builder /tmp/build/server /app/server

EXPOSE 7000

CMD [ "/app/server" ]
