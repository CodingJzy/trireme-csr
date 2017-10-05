FROM ubuntu
MAINTAINER Bernard Van De Walle <bernard@aporeto.com>

RUN mkdir -p /opt/trireme
RUN apt-get update \
&& rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ADD trireme-csr /opt/trireme/trireme-csr

WORKDIR /opt/trireme
CMD ["./trireme-csr"]
