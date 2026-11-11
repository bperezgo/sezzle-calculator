# syntax=docker/dockerfile:1

# Builder matches go.mod; it runs on the build host and cross-compiles for the target platform.
FROM --platform=$BUILDPLATFORM golang:1.27.0-alpine@sha256:4c9fe60190a2a3350ddc51de80d0224b8a6698d12bdfc999fee45ea9d6c46dbc AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd ./cmd
COPY internal ./internal
ARG TARGETOS TARGETARCH VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/calculator ./cmd

# Distroless static: CA certificates and tzdata, uid 65532, no shell or package manager.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab
COPY --from=build /out/calculator /calculator
ENV PORT=8080
EXPOSE 8080
USER 65532:65532
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD ["/calculator", "-healthcheck"]
ENTRYPOINT ["/calculator"]
