FROM mariadb:10.3.10-bionic

COPY scripts/* /docker-entrypoint-initdb.d
COPY mariadb/docker-entrypoint.sh /usr/local/bin/
