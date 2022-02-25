FROM nginx:latest

LABEL name="OPTISAM" \
    description="Optisam Meta Service" 

COPY meta-service/nginx.conf /etc/nginx/nginx.conf

WORKDIR /usr/share/nginx/html
COPY meta-service/version.html .
COPY meta-service/about.json .
