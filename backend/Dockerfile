FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/runplanner ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /bin/runplanner .
COPY db/migrations ./db/migrations
RUN mkdir -p data
EXPOSE 8080
CMD ["./runplanner"]
