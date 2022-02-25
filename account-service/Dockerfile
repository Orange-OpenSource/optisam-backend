FROM ubuntu:latest

LABEL name="OPTISAM" \
description="Optisam Account Service"

WORKDIR /opt/application/
COPY account_service/bin/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server
