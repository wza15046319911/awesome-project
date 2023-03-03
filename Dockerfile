FROM golang:alpine as build
#ENV
WORKDIR /opt
COPY . /opt
RUN go build .

FROM alpine as prod
WORKDIR /awesome
COPY --from=build /opt/awesome .
RUN mkdir -p /var/log/awesome

EXPOSE 8004
ENTRYPOINT ["./awesome"]
