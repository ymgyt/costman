FROM golang:1.11.2-alpine3.8 as build

RUN apk --no-cache add ca-certificates

WORKDIR /go/src/github.com/ymgyt/costman

COPY . ./

RUN CGO_ENABLED=0 go build -o /costman


FROM alpine:3.8

WORKDIR /root

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /costman .

ENTRYPOINT ["./costman"]