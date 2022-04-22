FROM golang:1.17-alpine3.15 AS builder

RUN apk --no-cache add \
    --virtual .build-deps \
    	alpine-sdk \
    	cmake \
    	sudo \
    	libssh2 libssh2-dev \
    	git \
    	dep \
    	bash \
    	curl \
    imagemagick \
    imagemagick-dev

WORKDIR /go/src/app
COPY go.* ./
RUN go mod download
COPY *.go ./
RUN go build -o app .

FROM alpine:3.15

RUN apk --no-cache add imagemagick
COPY --from=builder /go/src/app /app
RUN addgroup -g 1000 -S app && \
    adduser -u 1000 -G app -S app
USER app
ENTRYPOINT ["/app"]
