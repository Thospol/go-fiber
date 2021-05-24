FROM golang:1.16.4-alpine3.13 as builder
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk update && apk upgrade && \
  apk add --no-cache ca-certificates git

ENV TZ=Asia/Bangkok
RUN mkdir /api
WORKDIR /api
ADD . /api
RUN go get -u github.com/swaggo/swag/cmd/swag@v1.6.7
RUN swag init
RUN go mod download
RUN go build -o api

FROM alpine:3.12.0

RUN apk update && apk upgrade && \
  apk add --no-cache ca-certificates tzdata

COPY --from=builder /api/api .
COPY --from=builder /api/docs /docs

ADD /configs /configs

ENV TZ=Asia/Bangkok

EXPOSE 8000

CMD ["./api","-environment", "prod"]
