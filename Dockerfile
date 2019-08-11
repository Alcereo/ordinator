FROM golang:alpine as builder
RUN mkdir /build
RUN apk add git

WORKDIR /build
ADD ./go.mod /build/go.mod
RUN go build || true

ADD . /build/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .
FROM scratch
COPY --from=builder /build/main /app/
COPY --from=builder /build/config.yaml /app/
WORKDIR /app
CMD ["./main"]