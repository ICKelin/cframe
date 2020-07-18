FROM ubuntu:16.04
ADD dist/controller /
ADD dist/controller.toml /
RUN chmod +x /controller
CMD ["/controller", "-c", "/controller.toml"]