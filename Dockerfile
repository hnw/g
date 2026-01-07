FROM gcr.io/distroless/static:nonroot

ARG TARGETARCH

COPY bin/linux-${TARGETARCH}/g /usr/local/bin/g

ENTRYPOINT ["g"]
