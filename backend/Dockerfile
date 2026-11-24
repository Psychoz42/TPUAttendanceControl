FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /tpu_att .

FROM alpine:3.22
RUN adduser -S -u 1001 app
WORKDIR /app
COPY --from=build /tpu_att /usr/local/bin/tpu_att
USER app
ENV LISTEN_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["tpu_att"]
