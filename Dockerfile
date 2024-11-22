# syntax=docker/dockerfile:1
# sudo docker run --name "Domofonred" -d -v ./auth.json:/go/auth.json -v ./gate.json:/go/gate.json  bot_red

FROM golang:1.23

COPY ./ ./

RUN go env -w GO111MODULE=on

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot_red

EXPOSE 80

# Run
CMD ["/bot_red"]