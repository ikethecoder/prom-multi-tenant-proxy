FROM golang:1.15.2

WORKDIR /go/src/app

COPY cmd ./cmd
COPY internal ./internal
COPY go.* Makefile ./

RUN make go-build go-install

CMD ["bin/app"]
