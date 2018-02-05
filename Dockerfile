FROM golang:1.9.3-alpine3.7

WORKDIR /go/src/app

COPY . .

ENV LABEL_SELECTOR="k8s-iprouted/routeable=true"

ENV SUBNET="192.168.0.0/24"

RUN apk --no-cache add --virtual .build-dependencies \
    curl \
    git \
    && curl https://glide.sh/get | sh \
    && glide i \
    && go build -i -o $GOPATH/bin/k8s-iprouted \
    && apk del .build-dependencies \
    && rm -rf vendor && rm -rf ~/.glide

CMD k8s-iprouted $LABEL_SELECTOR $SUBNET

