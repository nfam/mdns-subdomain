FROM golang:alpine as builder
COPY . $GOPATH/src/mdns-subdomain
WORKDIR $GOPATH/src/mdns-subdomain
RUN CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' .

FROM scratch
COPY --from=builder /go/src/mdns-subdomain/mdns-subdomain /mdns-subdomain
EXPOSE 5353/udp
ENTRYPOINT ["/mdns-subdomain"]