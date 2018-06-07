FROM golang:onbuild as build_image
RUN mkdir /app
RUN mkdir /go/src/app
ADD . /go/src/app
WORKDIR /go/src/app
RUN go get -u github.com/golang/dep/...
RUN dep ensure
RUN GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=build_image /app/main .
CMD ["./main"]
