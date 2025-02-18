FROM spaulg/php-dev-containers:8.4

RUN echo "deb http://deb.debian.org/debian bookworm-backports main" >> /etc/apt/sources.list.d/bookworm-backports.list
RUN apt update -y
RUN apt install -y golang-1.23 ca-certificates
RUN ln -s /usr/lib/go-1.23/bin/go /usr/local/bin/go
