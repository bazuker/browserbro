FROM golang:1.22.3-alpine3.19 as build

RUN mkdir -p /go/src/browserbro

WORKDIR /go/src/browserbro

COPY . /go/src/browserbro

RUN go build -o browserbro -ldflags "-X main.version=1.0.0 -X 'main.date=$(date)'"

FROM alpine:3.19.1

RUN mkdir -p /opt/browserbro
COPY --from=build /go/src/browserbro/browserbro /opt/browserbro/browserbro
ENTRYPOINT ["/opt/browserbro/browserbro"]