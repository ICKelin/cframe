FROM ubuntu:16.04
RUN apt update && apt install -y iproute2 && apt install -y net-tools
COPY edge /
COPY config.toml /
RUN mkdir /log
ENTRYPOINT ["/edge","-c"]
CMD ["/config.toml"]