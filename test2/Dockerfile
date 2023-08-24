# syntax=docker/dockerfile:1

FROM golang:1.20.6-bookworm@sha256:b6c53162b13ec2ac1c725078671dbff289d9e723c045c27d73eacf0a50893598
WORKDIR /app
RUN apt update && apt install -y curl librocksdb7.8 librocksdb-dev
COPY . .
RUN go mod download
RUN go test && rm -rf db
RUN go build -o main main.go
EXPOSE 8080
CMD [ "/bin/bash" ]
# CMD [ "/app/main" ]
