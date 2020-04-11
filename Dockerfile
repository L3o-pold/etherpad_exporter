FROM golang:1.13-buster as build
MAINTAINER leopold.jacquot@gmail.com

WORKDIR /go/src/app
COPY . .

RUN make

FROM alpine as app

COPY --from=build /go/src/app/etherpad_exporter /bin/etherpad_exporter

ENTRYPOINT ["/bin/etherpad_exporter"]
EXPOSE 9301