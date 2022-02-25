FROM ubuntu:latest
LABEL name="OPISAM" \
    description="Optisam License Service"

WORKDIR /opt/application/
COPY license_service/bin/ /opt/application/
COPY license-service/cmd/server/rbac.rego/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server