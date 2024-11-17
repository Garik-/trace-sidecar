FROM golang:1.23

WORKDIR /go/src/github.com/Garik-/trace-sidecar

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG BUILD_VERSION=""
RUN CGO_ENABLED=0 GOOS=linux BUILD_VERSION=${BUILD_VERSION} make build

FROM scratch
COPY --from=0 /go/src/github.com/Garik-/trace-sidecar/bin/trace-sidecar .
ENTRYPOINT ["/trace-sidecar"]