FROM ubuntu:latest
LABEL name="OPISAM" \
    description="Optisam Report Service"

WORKDIR /opt/application/
COPY report_service/bin/ /opt/application/
COPY report-service/cmd/server/rbac.rego/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server