# builder image
ARG ARCH
FROM golang:1.16 as builder
ARG ARCH

WORKDIR /github.com/toppr-systems/dops

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make build.$ARCH

# final image
FROM $ARCH/alpine:3.14

COPY --from=builder /github.com/toppr-systems/dops/build/dops /bin/dops

USER 65534

ENTRYPOINT ["/bin/dops"]
