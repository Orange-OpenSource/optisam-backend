FROM ubuntu:latest
LABEL name="OPISAM" \
    description="Optisam Equipment Service"

WORKDIR /opt/application/
COPY equipment_service/bin/ /opt/application/
COPY equipment-service/cmd/server/rbac.rego/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server