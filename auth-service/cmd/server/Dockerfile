FROM ubuntu

WORKDIR /usr/local/bin

ADD server .
ADD key.pem .
ADD cert.pem .
ADD config-local.toml .

CMD [ "./server" ]