# Run the build
FROM golang:alpine
ENV WORKDIR /go/src/github.com/mojlighetsministeriet/identity-provider
COPY . $WORKDIR
WORKDIR $WORKDIR
RUN apk --update add git
RUN go get -t -v ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

# Create the final docker image
FROM scratch
COPY --from=0 /usr/local/go/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
COPY --from=0 /go/src/github.com/mojlighetsministeriet/identity-provider/identity-provider /
ENV RSA_PRIVATE_KEY ""
ENV DATABASE_TYPE "mysql"
ENV DATABASE "user:password@/dbname?charset=utf8mb4,utf8&parseTime=True&loc=Europe/Stockholm"
ENTRYPOINT ["/identity-provider"]
