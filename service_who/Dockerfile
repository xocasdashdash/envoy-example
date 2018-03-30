FROM golang:1.9 as oven
WORKDIR /go/src/github.com/xocasdashdash/envoy-example/service_who
ADD ./vendor ./vendor/
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -o bin/who && chmod +x bin/who
FROM scratch
COPY --from=oven /go/src/github.com/xocasdashdash/envoy-example/service_who/bin/who /app
EXPOSE 3333
ENTRYPOINT [ "/app" ]
