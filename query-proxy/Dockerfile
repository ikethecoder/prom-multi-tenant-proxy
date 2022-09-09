FROM golang:1.18.6

WORKDIR /go/src/app

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
COPY go.mod Makefile ./

RUN make go-tidy go-build go-install

CMD ["bin/app"]
