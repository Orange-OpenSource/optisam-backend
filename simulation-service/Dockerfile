FROM ubuntu:latest

LABEL name="OPTISAM" \
    description="Optisam Simulation Service" 

WORKDIR /opt/application/
COPY simulation_service/bin/ /opt/application/
COPY simulation-service/cmd/server/rbac.rego/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server