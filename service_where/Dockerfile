FROM golang:1.9 as oven
WORKDIR /go/src/github.com/xocasdashdash/envoy-test/service_where
ADD ./vendor ./vendor/
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -o bin/where && chmod +x bin/where
FROM scratch
COPY --from=oven /go/src/github.com/xocasdashdash/envoy-test/service_where/bin/where /app
EXPOSE 3333
ENTRYPOINT [ "/app" ]
