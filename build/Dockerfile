FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/noobaa-operator \
    USER_UID=1001 \
    USER_NAME=noobaa-operator

# tar is needed for kubectl cp
RUN microdnf install tar

# install operator binary
COPY build/_output/bin/noobaa-operator ${OPERATOR}

# copy scripts
COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup
USER ${USER_UID}

ENTRYPOINT ["/usr/local/bin/entrypoint"]
CMD ["operator", "run"]
