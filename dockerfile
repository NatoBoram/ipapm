FROM golang AS build
WORKDIR /go/src/ipapm

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go install -trimpath ./...

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /go/bin /usr/local/bin

ENTRYPOINT ["/usr/local/bin/ipapm"]
