FROM golang:latest AS builder
WORKDIR /workspace
ENV CGO_ENABLED=0
RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache
COPY ./go.* ./
RUN --mount=type=cache,target=/gomod-cache \
  go mod download
COPY . ./
ARG VERSION=1.0.0
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
  go build -ldflags "-X main.Version=${VERSION}" -o app .

FROM alpine:latest
RUN apk update && apk upgrade
RUN apk add --no-cache tzdata
ENV PUID=1000
ENV PGID=1000
ENV TZ="Europe/London"
RUN addgroup -g "${PGID}" -S appusers
RUN adduser -S -D -H -h /bin -u "${PUID}" -g "${PGID}" appuser
COPY --from=builder /workspace/app /bin
COPY --from=builder /workspace/brands.yml /bin

USER ${PUID}:${PGID}
WORKDIR /bin
CMD ["/bin/app"]
VOLUME /data
