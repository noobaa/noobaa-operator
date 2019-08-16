FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/noobaa-operator \
    CSV_TEMPLATES=/etc/noobaa-csv-templates \
    USER_UID=1001 \
    USER_NAME=noobaa-operator

# install operator binary
COPY build/_output/bin/noobaa-operator ${OPERATOR}

# copy csv template files
COPY build/_output/templates ${CSV_TEMPLATES}

# copy scripts
COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup
USER ${USER_UID}

ENTRYPOINT ["/usr/local/bin/entrypoint"]
CMD ["operator", "run"]
