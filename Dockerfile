FROM gcr.io/distroless/base@sha256:44a80dc28b49f47396e72ab2dc644b3f849880335de09625762faed207bb891d
LABEL \
        org.opencontainers.image.title="Firebolt OpenTelemetry Exporter" \
        org.opencontainers.image.description="Image for Firebolt OpenTelemetry Exporter" \
        org.opencontainers.image.vendor="Firebolt" \
        org.opencontainers.image.authors="Services Team" \
        org.opencontainers.image.source="https://github.com/firebolt-db/otel-exporter"
USER nonroot:nonroot
COPY ./otel-exporter /otel-exporter
ENTRYPOINT ["/otel-exporter"]
