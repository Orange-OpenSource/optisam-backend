FROM ubuntu:latest
LABEL name="OPISAM" \
    description="Optisam Application Service" 

WORKDIR /opt/application/
COPY application_service/bin/ /opt/application/
COPY application-service/cmd/server/rbac.rego/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server
