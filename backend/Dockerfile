FROM golang:1.22-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/seed-server  ./cmd/server  && \
    go build -o /bin/seed-migrate ./cmd/migrate && \
    go build -o /bin/seed-seed    ./cmd/seed

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /bin/seed-server  /usr/local/bin/seed-server
COPY --from=build /bin/seed-migrate /usr/local/bin/seed-migrate
COPY --from=build /bin/seed-seed    /usr/local/bin/seed-seed
COPY --from=build /app/migrations   /app/migrations
EXPOSE 8080
CMD ["seed-server"]
