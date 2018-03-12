FROM golang:1.9 as oven
WORKDIR /go/src/github.com/xocasdashdash/envoy-test/service_hello
ADD ./vendor ./vendor/
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -o bin/hello && chmod +x bin/hello
FROM scratch
COPY --from=oven /go/src/github.com/xocasdashdash/envoy-test/service_hello/bin/hello /app
EXPOSE 3333
ENTRYPOINT [ "/app" ]
