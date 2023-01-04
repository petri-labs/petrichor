FROM golang:1.18.7-bullseye AS build

RUN apt update && apt install jq build-essential -y

WORKDIR /petrichord
COPY . /petrichord

RUN curl https://get.ignite.com/cli@v0.25.1! | bash
RUN ignite chain build --output . --verbose

FROM debian:bullseye AS run
COPY ./scripts/containers/build-wrapper.sh /petrichord/build-wrapper.sh

VOLUME /petrichord
COPY --from=build /petrichord/ /petrichord/
WORKDIR /petrichord

EXPOSE 26656 26657
ENTRYPOINT ["/petrichord/build-wrapper.sh"]
CMD ["start", "--log_format", "plain"]
STOPSIGNAL SIGTERM
