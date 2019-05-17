FROM golang:1.12 AS builder
RUN go version
WORKDIR /go/src/github.com/aware-hq/azure-key-vault-controller/
COPY . .

RUN set -x && \
    go get github.com/golang/dep/cmd/dep
#    go get -u honnef.co/go/tools/cmd/staticcheck
RUN dep ensure -v
# RUN staticcheck ...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o azure-key-vault-controller .

FROM registry.access.redhat.com/ubi7-dev-preview/ubi-minimal:7.6

ENV OPERATOR=/usr/local/bin/azure-key-vault-controller \
    USER_UID=1001 \
    USER_NAME=azure-key-vault-controller

# install operator binary
COPY --from=builder /go/src/github.com/aware-hq/azure-key-vault-controller/azure-key-vault-controller ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
