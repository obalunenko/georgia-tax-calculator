FROM ghcr.io/obalunenko/go-tools:v0.13.0 AS builder

# Configure least privilege user
ARG UID=2000
ARG GID=2000
RUN addgroup -S gotools -g ${UID} && \
    adduser -S gotools -u ${GID} -G gotools -h /home/gotools -s /bin/sh -D gotools

CMD ["/bin/sh", "-c", ""]

USER gotools