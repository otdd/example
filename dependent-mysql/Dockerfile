FROM mysql:8.0.19
MAINTAINER me

ENV MYSQL_DATABASE=hello_otdd \
    MYSQL_ROOT_PASSWORD=123456

ADD schema.sql /docker-entrypoint-initdb.d

EXPOSE 3306
