FROM golang:1.12.16 as build
RUN mkdir -p /app/building
WORKDIR /app/building
ADD . /app/building
RUN make build

FROM centos:7.7.1908
COPY --from=build /app/building/dist/bin/discovery /app/bin/
COPY --from=build /app/building/dist/conf/discovery.toml /app/conf/
ENV  LOG_DIR    /app/logs
WORKDIR /app/
CMD  /app/bin/discovery -conf /app/conf/ -confkey discovery.toml
