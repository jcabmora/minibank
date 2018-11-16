FROM mariadb:10.3.10-bionic

RUN apt-get update \
    && apt-get install -y \
        python3 \
        python3-distutils \
        wget \
    && wget 'https://bootstrap.pypa.io/get-pip.py' \
    && python3 get-pip.py \
    && rm -rf /var/lib/apt/lists/*

COPY mariadb/requirements.txt \
    mariadb/clustering.py ./
RUN pip install -r requirements.txt

COPY scripts/* /docker-entrypoint-initdb.d
COPY mariadb/docker-entrypoint.sh /usr/local/bin
