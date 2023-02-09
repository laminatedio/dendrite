FROM gcr.io/distroless/base-debian11
COPY dendrite /
ENTRYPOINT ["/dendrite", "server"]
