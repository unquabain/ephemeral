FROM golang:1.21 AS BUILD

ENV CGO_ENABLED=0
RUN mkdir /build
WORKDIR /build
COPY . .
RUN go mod tidy
RUN go build -o ephemeral

FROM scratch
COPY --from=BUILD /build/ephemeral ephemeral
EXPOSE 8989
ENTRYPOINT ["/ephemeral", "serve"]
