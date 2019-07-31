FROM golang as builder
WORKDIR /
RUN git clone https://github.com/aasmall/dicemagic
WORKDIR /dicemagic
RUN go mod download
WORKDIR /dicemagic/dice-server
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -v -installsuffix cgo -o /dicemagic/bin/dice-server .

FROM alpine:latest AS dice-server
RUN apk --no-cache add ca-certificates
COPY --from=builder /dicemagic/bin/dice-server /dicemagic/
COPY dice-server/.env /dicemagic/
COPY dicemagic-afb0a5ec5454.json /dicemagic/
# ENTRYPOINT [ "/bin/sh", "-c", "cd /dicemagic && . /dicemagic/.env && ./dice-server" ]

EXPOSE 8080