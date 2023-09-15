FROM golang:latest as builder

WORKDIR /app
COPY main.go ./

RUN go build -o main main.go


FROM ubuntu:latest

COPY --from=builder /app/main /app/main

WORKDIR /


EXPOSE 9999
ENV PORT=9999

CMD ["/app/main"]


