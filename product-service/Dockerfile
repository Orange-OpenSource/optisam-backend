FROM ubuntu:latest
LABEL name="OPISAM" \
    description="Optisam Product Service" 

WORKDIR /opt/application/
COPY product_service/bin/ /opt/application/
COPY product-service/cmd/server/rbac.rego/ /opt/application/
EXPOSE 8080
ENTRYPOINT ./server