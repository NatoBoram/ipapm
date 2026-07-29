FROM golang AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /app/ipapm /ipapm

ENTRYPOINT ["/ipapm"]
