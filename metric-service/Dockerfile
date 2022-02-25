FROM ubuntu:latest
LABEL name="OPISAM" \
    description="Optisam Metric Service" 

WORKDIR /opt/application/
COPY metric_service/bin/ /opt/application/
COPY metric-service/cmd/server/rbac.rego/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server