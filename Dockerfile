FROM golang:1.4

RUN go get -v github.com/constabulary/gb/...

WORKDIR /usr/src/gdbuild
ENV PATH /usr/src/gdbuild/bin:$PATH
COPY . /usr/src/gdbuild

RUN gb build
