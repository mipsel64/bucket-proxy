# distroless/static ships CA certificates and tzdata; the binary is built with CGO_ENABLED=0.
ARG BASE_IMAGE=gcr.io/distroless/static-debian13:nonroot
ARG BUILDER_IMAGE=docker.io/library/golang:1.27.4

FROM ${BUILDER_IMAGE} AS builder

ARG TARGETOS
ARG TARGETARCH
ARG NOOPT
ARG NOSTRIP
ARG VERSION
ARG GIT_COMMIT

WORKDIR /go/src/bucket-proxy
RUN --mount=type=bind,readwrite,target=/go/src/bucket-proxy --mount=target=/root/.cache,type=cache --mount=target=/go/pkg,type=cache \
  make GOARCH=${TARGETARCH} GOOS=${TARGETOS} \
  VERSION=${VERSION} GIT_COMMIT=${GIT_COMMIT} NOOPT=${NOOPT} NOSTRIP=${NOSTRIP} \
  DESTDIR=/tmp/install/${TARGETOS}/${TARGETARCH} install

FROM ${BASE_IMAGE} AS release
ARG TARGETOS
ARG TARGETARCH

LABEL maintainer="vietanhs0817@gmail.com"
LABEL org.opencontainers.image.source=https://github.com/vietanhduong/bucket-proxy
LABEL org.opencontainers.image.description="A simple bucke proxy server"
LABEL org.opencontainers.image.licenses=MIT

COPY --from=builder /tmp/install/${TARGETOS}/${TARGETARCH}/usr/bin/* /usr/bin/

ENTRYPOINT [ "/usr/bin/bucket-proxy" ]
