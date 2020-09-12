FROM centos:7.5.1804

COPY api-service /
COPY config.toml /
COPY start.sh /
RUN chmod +x start.sh && chmod +x api-service
RUN mkdir /log
CMD /start.sh
