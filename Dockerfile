FROM golang:1.5-alpine

RUN apk add --update git && rm -rf /var/cache/apk/*

RUN go get -v github.com/constabulary/gb/...

WORKDIR /usr/src/gdbuild
ENV PATH /usr/src/gdbuild/bin:$PATH
COPY . /usr/src/gdbuild

RUN gb build
RUN GOPATH="$PWD:$PWD/vendor" go build -a -installsuffix netgo -tags netgo -ldflags '-d' -a -o bin/gdbuild-static ./src/cmd/gdbuild
