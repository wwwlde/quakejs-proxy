FROM scratch
ARG TARGETARCH
COPY --chmod=755 ${TARGETARCH}/quakejs-proxy /quakejs-proxy
ENTRYPOINT ["/quakejs-proxy"]
