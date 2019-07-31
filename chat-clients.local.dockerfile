FROM golang as builder
WORKDIR /
RUN git clone https://github.com/aasmall/dicemagic
WORKDIR /dicemagic
RUN go mod download
WORKDIR /dicemagic/chat-clients
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -v -installsuffix cgo -o /dicemagic/bin/chat-clients .

FROM alpine:latest AS chat-clients
RUN apk --no-cache add ca-certificates redis
COPY --from=builder /dicemagic/bin/chat-clients /dicemagic/
COPY chat-clients/.env /dicemagic/
COPY dicemagic-afb0a5ec5454.json /dicemagic/
# ENTRYPOINT [ "/bin/sh", "-c", " redis-server & sleep 2 && cd /dicemagic && . /dicemagic/.env && ./chat-clients & /bin/sh" ]
EXPOSE 7070