FROM golang:1.15 AS build

WORKDIR /src
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=readonly -ldflags="-w -s" -o /bin/service

FROM scratch
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/service /bin/service
ENTRYPOINT ["/bin/service"]