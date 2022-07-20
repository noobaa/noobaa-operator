FROM python:3.9-slim
RUN pip install --no-cache-dir pykmip
WORKDIR /work
COPY start.sh         start.sh
COPY server.conf      server.conf
COPY certs/ca.crt     ca.cert
COPY certs/server.crt server.cert
COPY certs/server.key server.key

CMD /work/start.sh
