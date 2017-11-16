FROM golang:1.7.3
WORKDIR /go/src/github.com/syncthing/
RUN git clone https://github.com/mdevey/syncthing.git -b stserve
WORKDIR /go/src/github.com/syncthing/syncthing
RUN CGO_ENABLED=0 GOOS=linux ./build.sh

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN adduser -D syncthing
RUN mkdir -p /home/syncthing/.config/ && chmod 777 /home/syncthing/.config
USER syncthing
ENV HOME /home/syncthing
RUN ln -s /var/syncthing/config /home/syncthing/.config/syncthing
EXPOSE 8765
COPY --from=0 /go/src/github.com/syncthing/syncthing/bin/stserve .
CMD ["./stserve"]

