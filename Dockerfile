FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/perimatrik-api ./cmd/server

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/perimatrik-api /app/perimatrik-api
ENV PORT=8000 STORAGE_PATH=/app/storage/app/public
VOLUME ["/app/storage/app/public"]
EXPOSE 8000
CMD ["/app/perimatrik-api"]
