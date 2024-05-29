FROM gcr.io/distroless/base@sha256:786007f631d22e8a1a5084c5b177352d9dcac24b1e8c815187750f70b24a9fc6
LABEL \
        org.opencontainers.image.title="Firebolt OpenTelemetry Exporter" \
        org.opencontainers.image.description="Image for Firebolt OpenTelemetry Exporter" \
        org.opencontainers.image.vendor="Firebolt" \
        org.opencontainers.image.authors="Services Team" \
        org.opencontainers.image.source="https://github.com/firebolt-db/otel-exporter"
USER nonroot:nonroot
COPY ./otel-exporter /otel-exporter
ENTRYPOINT ["/otel-exporter"]
