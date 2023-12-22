FROM --platform=$BUILDPLATFORM golang:alpine AS builder

RUN apk update && apk add --no-cache git
WORKDIR /go/src/media-host
COPY . .

ARG TARGETOS TARGETARCH
ENV GOOS $TARGETOS
ENV GOARCH $TARGETARCH
RUN go build -o /go/bin/media-host cmd/main.go

FROM scratch
COPY --from=builder /go/bin/media-host /go/bin/media-host
ENTRYPOINT ["/go/bin/media-host"]