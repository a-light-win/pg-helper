#FROM docker.io/library/debian:12
#RUN useradd -m --uid 65532 nonroot
FROM gcr.io/distroless/static-debian12:nonroot
COPY ./dist/pg-helper /usr/bin/pg-helper
CMD ["pg-helper", "serve"]
