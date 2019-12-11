FROM ubuntu

WORKDIR /usr/local/bin

ADD server .
ADD cert.pem .
ADD config-dev.toml .

CMD [ "./server" ]