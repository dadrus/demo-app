# Builder image to build the app
FROM golang:1.14-buster as builder
LABEL maintainer=dimitrij.drus@innoq.com

RUN apt-get update && apt-get install -y xz-utils

# UPX is GPL
ADD https://github.com/upx/upx/releases/download/v3.94/upx-3.94-amd64_linux.tar.xz /usr/local
RUN xz -d -c /usr/local/upx-3.94-amd64_linux.tar.xz | \
    tar -xOf - upx-3.94-amd64_linux/upx > /bin/upx && \
    chmod a+x /bin/upx

ENV USER=appuser
ENV UID=10001

RUN adduser \
    --disabled-login \
    --gecos "" \
    --home "/nonexistent" \
    --no-create-home \
    --shell "/sbin/nologin" \
    --uid "${UID}" \
    "${USER}"

WORKDIR /go/src/demo-app

ADD . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"
RUN strip --strip-unneeded demo-app
RUN upx demo-app

# The actual image of the app
FROM scratch
LABEL maintainer=dimitrij.drus@innoq.com

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /go/src/demo-app/demo-app /opt/demo-app/demo-app
COPY web /opt/demo-app/web

WORKDIR /opt/demo-app

USER appuser:appuser

ENV GIN_MODE=release
ENV PORT 8082

EXPOSE $PORT
ENTRYPOINT ["/opt/demo-app/demo-app"]
