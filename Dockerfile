FROM golang:1.15-buster as build

WORKDIR /workspace
COPY . /workspace

RUN make build

FROM debian:buster-slim

RUN apt-get -y update && apt-get install -y ca-certificates

WORKDIR /sats-stacker

COPY --from=build /workspace/sats-stacker /sats-stacker

USER nobody
CMD ["/sats-stacker/sats-stacker"]
