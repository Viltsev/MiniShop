FROM golang

RUN mkdir /app
ADD . /app
WORKDIR /app

COPY wait-for-it.sh .
RUN chmod +x wait-for-it.sh

RUN go build -o main ./cmd/server/main.go

EXPOSE 8083

CMD ["./main"]