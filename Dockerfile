FROM golang:1.22.6-alpine
WORKDIR /bdPractice
COPY . .
RUN go mod download
RUN go build -o main .
EXPOSE 8080
CMD ["./main"]
