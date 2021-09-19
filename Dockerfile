FROM golang:1.16-buster as builder

WORKDIR /app

COPY ./config.json .
COPY ./main.go .

RUN go build main.go

RUN rm main.go

FROM gcr.io/distroless/base-debian10

WORKDIR /app

COPY --from=builder /app .

ENV HOST 0.0.0.0
EXPOSE 9999

USER nonroot:nonroot

CMD [ "./main" ]
