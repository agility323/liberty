FROM golang:1.18-buster
MAINTAINER zhangbeihai <zhangbeihai@funplus.com>

COPY docker_deb/ docker_deb/

RUN cat /etc/issue \
	&& apt-get update -y \
	&& apt-get install -y iptables \
	&& apt-get install -y libdevmapper1.02.1 \
	&& go version \
	&& cd docker_deb \
	&& pwd \
	&& ls -alh \
	&& dpkg -i containerd.io_1.6.4-1_amd64.deb \
	&& dpkg -i docker-ce-cli_20.10.9~3-0~debian-buster_amd64.deb \
	&& dpkg -i docker-ce_20.10.9~3-0~debian-buster_amd64.deb
