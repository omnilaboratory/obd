FROM ubuntu

RUN apt-get update && apt-get install -y python3.9 python3-pip

WORKDIR /test
COPY  reqs.txt .
RUN python3.9 -m pip install -r reqs.txt

COPY omnibolt.py .
COPY omnicore_connection.py .
COPY runner.py .
COPY test.py .
COPY bitcoin.conf .
COPY conf.tracker.ini .

ENV OMNI_BOLT_ALICE=omnibolt_alice:60020
ENV OMNI_BOLT_BOB=omnibolt_bob:60020
ENV BITCOIN_CONF=/test/bitcoin.conf

ENTRYPOINT [ "python3.9", "-m", "pytest", "test.py" ]
