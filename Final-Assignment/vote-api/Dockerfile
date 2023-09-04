FROM golang:1.20 AS build-stage

WORKDIR /

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /vote-api

FROM alpine:latest AS run-stage

WORKDIR /

COPY --from=build-stage /vote-api /vote-api

EXPOSE 1080

ENV REDIS_URL=host.docker.internal:6379

CMD ["/vote-api"]



