FROM cgr.dev/chainguard/static

ARG TARGETPLATFORM

COPY $TARGETPLATFORM/osminokal /usr/bin/

ENTRYPOINT ["/usr/bin/osminokal"]
