FROM ubuntu

WORKDIR /usr/local/bin

ADD server .
ADD config.toml .
ADD config-dev.toml .
ADD config-int.toml .
ADD cert.pem .

CMD [ "./server" ]