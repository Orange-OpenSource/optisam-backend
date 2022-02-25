FROM ubuntu:latest

LABEL name="OPTISAM" \
    description="Optisam Import Service" 

WORKDIR /opt/application/
COPY import_service/bin/ /opt/application/
COPY import-service/cmd/server/rbac.rego/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server