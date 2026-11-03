FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY cmd ./cmd
RUN CGO_ENABLED=0 go build -o /out/migrator ./cmd/migrator

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/migrator /usr/local/bin/migrator
COPY migrations ./migrations
ENTRYPOINT ["migrator"]
