FROM golang:1.12 as builder
WORKDIR /go/src/github.com/iyacontrol/prophet
ADD ./  /go/src/github.com/iyacontrol/prophet
RUN CGO_ENABLED=0 go build


FROM geekidea/alpine-a:3.9
COPY --from=builder /go/src/github.com/iyacontrol/prophet/prophet /usr/local/bin/prophet
RUN chmod +x /usr/local/bin/prophet
CMD ["prophet"]