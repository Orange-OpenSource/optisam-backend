FROM ubuntu:latest
LABEL name="OPISAM" \
    description="Optisam DPS Service" 

WORKDIR /opt/application/
COPY dps_service/bin/ /opt/application/
COPY dps-service/cmd/server/rbac.rego/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server
