FROM node:8-alpine

RUN addgroup -g 2000 -S skia && \
    adduser -u 2000 -S skia -G skia && \
    apk update && \
    apk add --no-cache git make ca-certificates

COPY . /

USER skia

ENTRYPOINT ["/jsdoc"]
CMD ["--logtostderr", "--port=:8000", "--prom_port=:20000", "--git_repo_dir=/tmp"]
