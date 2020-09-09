FROM centos:7.5.1804

COPY apiserver /
COPY config.toml /
RUN mkdir /log
ENTRYPOINT ["/apiserver","-c"]
CMD ["/config.toml"]