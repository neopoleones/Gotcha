FROM golang:1.18

WORKDIR /app/gotcha
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8080
CMD ["gotcha-app"]