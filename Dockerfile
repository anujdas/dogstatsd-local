FROM golang:1.18.1 as build

WORKDIR /src

ARG GOOS=linux
ARG CGO_ENABLED=0

COPY . /src
RUN go build && go test

###

FROM scratch

COPY --from=build /src/dogstatsd-local .
EXPOSE 8125/udp

ENTRYPOINT ["/dogstatsd-local"]
CMD ["--port", "8125"]
