# Container image for the harrow CLI, following the gomatic/build `make docker`
# contract: the `docker` target runs `build-all` (goreleaser) first, so the
# binaries already exist under dist/ as <bin>-<os>-<arch>. This image only
# copies the right one in — it does not build from source. The build context is
# the repo root and must include dist/ (see .dockerignore). Build it with
# `make docker` (single-arch) or `make docker-buildx` (multi-arch manifest).
#
# ENTRYPOINT_BIN is the binary id; the Makefile passes DOCKER_ENTRYPOINT (the
# first builds: id from .goreleaser.yml). TARGETOS/TARGETARCH are provided by
# `docker build --platform`.
#
# The gomatic convention is `FROM gomatic/runtime` (non-root user, CA certs,
# distroless base). distroless/static:nonroot gives a minimal, non-root result;
# harrow links pg_query (cgo), so the release build must link it statically for
# this base — see .goreleaser.yml.
ARG ENTRYPOINT_BIN=harrow
FROM gcr.io/distroless/static:nonroot
ARG ENTRYPOINT_BIN
ARG TARGETOS
ARG TARGETARCH

# Copy the prebuilt binary to a fixed path so the exec-form ENTRYPOINT does not
# depend on the (build-arg) binary name.
COPY dist/${ENTRYPOINT_BIN}-${TARGETOS}-${TARGETARCH} /usr/local/bin/app
ENTRYPOINT ["/usr/local/bin/app"]
