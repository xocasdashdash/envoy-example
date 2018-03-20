FROM golang:1.9 as oven
WORKDIR /go/src/github.com/xocasdashdash/envoy-test/service_orchestra
ADD ./vendor ./vendor/
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -o bin/orchestra && chmod +x bin/orchestra
FROM scratch
COPY --from=oven /go/src/github.com/xocasdashdash/envoy-test/service_orchestra/bin/orchestra /app
EXPOSE 3333
ENTRYPOINT [ "/app" ]
