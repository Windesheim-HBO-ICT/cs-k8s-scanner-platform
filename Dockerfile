# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build

ARG TARGETOS
ARG TARGETARCH
ARG APP_VERSION=dev

WORKDIR /src
COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${APP_VERSION}" \
    -o /out/scanner-platform .

FROM scratch

ARG APP_VERSION=dev

LABEL org.opencontainers.image.title="Scanner Platform testapp" \
      org.opencontainers.image.description="Kubernetes testapp for Deployment, Service, ConfigMap, Secret and rolling update validation." \
      org.opencontainers.image.source="https://github.com/Windesheim-HBO-ICT/cs-k8s-scanner-platform" \
      org.opencontainers.image.version="${APP_VERSION}"

USER 65532:65532
COPY --from=build /out/scanner-platform /scanner-platform

EXPOSE 8080
ENTRYPOINT ["/scanner-platform"]
