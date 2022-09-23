FROM golang:alpine as wrapper

RUN mkdir /go/RestApi

WORKDIR /go/RestApi

COPY . .

RUN go build -tags=jsoniter -o rest

FROM alpine

COPY --from=wrapper /go/RestApi/rest /go/rest

CMD ["/go/rest"]