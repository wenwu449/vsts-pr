# build stage
FROM golang:alpine AS build-env

ADD . /go/src/github.com/wenwu449/vsts-pr

WORKDIR /go/src/github.com/wenwu449/vsts-pr/
RUN go test -v && CGO_ENABLED=0 GGOS=linux go build -o vsts-pr

# final stage
FROM alpine

RUN apk add --no-cache ca-certificates
WORKDIR /vsts-pr
COPY --from=build-env /go/src/github.com/wenwu449/vsts-pr/vsts-pr .

CMD ["./vsts-pr"]