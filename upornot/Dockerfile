FROM golang:1.12.5 AS build

WORKDIR /work
COPY *.go ./
ENV CGOENABLED=0
RUN go build -o upornot .

EXPOSE 4444
ENTRYPOINT ["/work/upornot", "-addr", "0.0.0.0:4444"]
