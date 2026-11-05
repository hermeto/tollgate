# Runtime image used by GoReleaser (dockers_v2). Binaries are prebuilt per platform.
# Base: distroless static, non-root, pinned by digest (linux/amd64 and linux/arm64).
FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/tollgate /usr/bin/tollgate
USER 65532:65532
ENTRYPOINT ["/usr/bin/tollgate"]
