FROM golang:bookworm

WORKDIR /app

COPY . .

RUN go mod tidy
RUN make test-server-build

CMD [ "make", "test-server-start"]