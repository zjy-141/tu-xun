FROM alpine:latest

RUN apk update && apk add --no-cache tzdata ca-certificates
COPY go_app /runtime/go_app
WORKDIR /runtime
ENTRYPOINT ["/runtime/go_app"]
