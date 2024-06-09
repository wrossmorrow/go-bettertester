FROM golang:1.22-alpine

WORKDIR /app

COPY ./ ./
RUN go mod download
RUN go build -o /echo
EXPOSE 8000
CMD [ "/echo" ]
