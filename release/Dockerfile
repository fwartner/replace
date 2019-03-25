FROM alpine

LABEL maintainer="Florian Wartner <florian@wartner.io>"

RUN apk --no-cache add --virtual .gocrond-deps \
        ca-certificates  \
        wget \
    && REPLACE_VERSION=1.0.0 \
    && wget -O /usr/local/bin/replace https://github.com/fwartner/replace/releases/download/$REPLACE_VERSION/gr-64-linux \
    && chmod +x /usr/local/bin/replace \
    && apk del .gocrond-deps

CMD ["replace"]
