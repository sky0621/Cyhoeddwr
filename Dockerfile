# step 1. compile
FROM "sky0621dhub/dockerfile-gowithdep" AS builder

COPY . /go/src/github.com/sky0621/Cyhoeddwr
WORKDIR /go/src/github.com/sky0621/Cyhoeddwr
RUN dep ensure
RUN go test ./...
RUN CGO_ENABLED=0 go build -o cyhoeddwr github.com/sky0621/Cyhoeddwr

# -----------------------------------------------------------------------------
# step 2. build
FROM scratch
COPY --from=builder /go/src/github.com/sky0621/Cyhoeddwr/ .
ENTRYPOINT [ "./cyhoeddwr" ]
EXPOSE 14080
