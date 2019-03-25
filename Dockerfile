FROM golang:alpine AS buildenv

LABEL maintainer="Florian Wartner <florian@wartner.io>"

COPY . /go/src/replace
WORKDIR /go/src/replace

RUN apk --no-cache add git \
    && go get \
    && go build \
    && chmod +x replace \
    && ./replace --version

FROM alpine
COPY --from=buildenv /go/src/replace/replace /usr/local/bin
CMD ["replace"]
