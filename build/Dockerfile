FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/noobaa-operator \
    USER_UID=1001 \
    USER_NAME=noobaa-operator
ARG NOOBAA_BIN_PATH=build/_output/bin/noobaa-operator

# tar is needed for kubectl cp
RUN microdnf install tar

# install operator binary
COPY ${NOOBAA_BIN_PATH} ${OPERATOR}

# copy scripts
COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup
USER ${USER_UID}

ENTRYPOINT ["/usr/local/bin/noobaa-operator"]
CMD ["operator", "run"]
