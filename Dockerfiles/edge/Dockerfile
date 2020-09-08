FROM centos:7.5.1804

COPY edge /
COPY config.toml /

ENTRYPOINT ["/edge","-c"]
CMD ["/config.toml"]
