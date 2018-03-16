FROM instrumentisto/glide:0.13.1-go1.10 as builder
RUN apk --update add gcc musl-dev
WORKDIR /go/src/app
COPY . .
RUN glide install
RUN go build -a app .

FROM alpine:3.4
RUN apk --update add imagemagick && \
rm -rf /var/cache/apk/*
WORKDIR /root/
COPY --from=builder /go/src/app/app .
CMD ["./app"]