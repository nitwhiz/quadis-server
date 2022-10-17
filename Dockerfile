FROM golang:1.19.2-alpine3.16 as builder

WORKDIR /app

COPY ./ /app

RUN CGO_ENABLED=0 go build -o ./build/server ./cmd/server

FROM alpine:3.16.2

WORKDIR /app

COPY --from=builder /app/build/server /app/server

EXPOSE 7000

ENV GIN_MODE=release

CMD [ "/app/server" ]
