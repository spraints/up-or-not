FROM golang:1.11.5

WORKDIR /go/src/github.com/spraints/up-or-not/
COPY . .
RUN env CGOENABLED=0 go build -o up-or-not .

FROM scratch
