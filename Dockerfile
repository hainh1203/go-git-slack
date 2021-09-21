FROM golang:1.16-buster as builder

WORKDIR /app

COPY ./main.go .

RUN go build main.go

RUN rm main.go

FROM ubuntu:20.04

RUN apt-get update && apt-get install -y curl && apt-get install -y python

RUN curl https://sdk.cloud.google.com > install.sh

RUN bash install.sh --disable-prompts

WORKDIR /app

COPY --from=builder /app .

ENV HOST 0.0.0.0
EXPOSE 8080

ENTRYPOINT [ "./main" ]
