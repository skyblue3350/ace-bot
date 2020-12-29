FROM --platform=$BUILDPLATFORM golang:1.15.6 as build

WORKDIR /app
ARG TARGETARCH
ENV GOARCH=$TARGETARCH

RUN apt-get update -y \
  && apt-get install ca-certificates -y \
  && update-ca-certificates

COPY go.mod /app

RUN go get -u all

COPY ./cmd /app

RUN go build

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/ace-bot /ace-bot

ENTRYPOINT ["/ace-bot"]
