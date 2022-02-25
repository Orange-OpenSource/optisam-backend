FROM ubuntu:latest
LABEL name="OPISAM" \
    description="Optisam Auth Service" 

WORKDIR /opt/application/
COPY auth_service/bin/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server
