FROM golang:1.13 as build_image
ADD . /go/src/app
WORKDIR /go/src/app
RUN go mod download
RUN GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN mkdir /app
WORKDIR /app
COPY --from=build_image /go/src/app/main .
CMD ["./main"]
