FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/github.com/jzandbergen/mockttp
COPY . .

RUN go mod tidy
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/mockttp

FROM scratch
COPY --from=builder /go/bin/mockttp /go/bin/mockttp
ENTRYPOINT ["/go/bin/mockttp"]
