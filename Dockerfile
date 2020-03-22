FROM golang:1.14.1 AS build
WORKDIR /go/src/github.com/xnyo/ugr
COPY . .
RUN GOOS=linux go build -a -tags netgo -ldflags '-linkmode external -extldflags -static' -o ugr *.go

FROM alpine:3 AS runtime
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /go/src/github.com/xnyo/ugr/ugr ./
CMD ["./ugr"]
