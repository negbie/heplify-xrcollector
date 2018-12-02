# docker build --no-cache -t negbie/heplify-xrcollector:latest .
# docker push negbie/heplify-xrcollector:latest

FROM golang:alpine as builder
RUN apk update && apk add --no-cache git
RUN go get -u github.com/negbie/sipparser
COPY . /heplify-xrcollector
WORKDIR /heplify-xrcollector
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w' -installsuffix cgo -o heplify-xrcollector .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /heplify-xrcollector/heplify-xrcollector .