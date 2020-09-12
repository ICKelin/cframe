FROM centos:7.5.1804

COPY controller /
COPY config.toml /
COPY start.sh /
RUN chmod +x start.sh && chmod +x controller
CMD /start.sh
