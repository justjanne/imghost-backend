FROM golang

RUN apt-get update && apt-get install -y --no-install-recommends \
imagemagick \
libmagickwand-dev

WORKDIR /go/src/app
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -o app .

ENTRYPOINT ["./app"]
