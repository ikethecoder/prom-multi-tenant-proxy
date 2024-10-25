FROM golang:1.23.2

WORKDIR /go/src/app

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY go.mod Makefile ./

RUN make go-tidy go-build go-install

CMD ["bin/app"]
