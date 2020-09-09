FROM centos:7.5.1804

COPY controller /
COPY config.toml /

ENTRYPOINT ["/controller","-c"]
CMD ["/config.toml"]