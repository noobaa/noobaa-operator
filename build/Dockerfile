FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/noobaa-operator \
    USER_UID=1001 \
    USER_NAME=noobaa-operator
ARG NOOBAA_BIN_PATH=build/_output/bin/noobaa-operator

# tar is needed for kubectl cp
RUN microdnf -y install tar

# install operator binary
COPY ${NOOBAA_BIN_PATH} ${OPERATOR}

USER ${USER_UID}

ENTRYPOINT ["/usr/local/bin/noobaa-operator"]
CMD ["operator", "run"]
