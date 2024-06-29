FROM golang:alpine

WORKDIR /

COPY surface.go .

RUN go build -o surface surface.go

CMD [". /surface"]