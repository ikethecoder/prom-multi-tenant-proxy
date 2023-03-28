FROM golang:1.18.10

WORKDIR /go/src/app

COPY cmd ./cmd
COPY internal ./internal
COPY go.* Makefile ./

RUN make go-build
RUN go mod vendor
RUN make go-install

CMD ["bin/app"]
