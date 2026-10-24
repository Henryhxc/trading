FROM golang:1.25-alpine AS build
WORKDIR /src
COPY apps/backend/go.mod apps/backend/go.sum ./
RUN go mod download
COPY apps/backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOMAXPROCS=2 go build -p 2 -trimpath -o /out/ux-api ./cmd/api

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/ux-api /app/ux-api
COPY db/migrations/ /app/db/migrations/
USER app
EXPOSE 8080
ENTRYPOINT ["/app/ux-api"]
