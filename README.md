[![pipeline status](https://gitlab.tech.orange/optisam/optisam-it/optisam-backend/badges/develop/pipeline.svg)](https://gitlab.forge.orange-labs.fr/OrangeMoney/optisam/optisam-backend/commits/develop) [![coverage report](https://gitlab.tech.orange/optisam/optisam-it/optisam-backend/badges/develop/coverage.svg)](https://gitlab.forge.orange-labs.fr/OrangeMoney/optisam/optisam-backend/commits/develop)

OpTISAM
=======

## Introduction

__OPTISAM__ (Optimized tool for inventive Software Asset Management) is a tool for the Software Asset Management Compliance Audit and Optimization Tool. This monorepo contains all the backend services namely:

- [account-service](account-service/dbdoc/README.md)
- [acqrights-service](acqrights-service/dbdoc/README.md)
- [application-service](application-service/dbdoc/README.md)
- auth-service
- [dps-service](dps-service/dbdoc/README.md)
- equipment-service
- import-service
- license-service
- metric-service
- [product-service](product-service/dbdoc/README.md)
- [report-service](report-service/dbdoc/README.md)
- [simulation-service](simulation-service/dbdoc/README.md)

## Quick start
### Download

```
$ git clone https://gitlab.tech.orange/optisam/optisam-it/optisam-backend.git
```

### Build

##### - Change configuration files
<em>Update values of config files **${service}/cmd/server/config-local.toml** as per your requirement</em>

* Building docker images for all micro-services

```
cd ${service-name}/cmd/server
docker build --pull -t optisam/${service-name}-service:latest -f Dockerfile .
docker push optisam/${service-name}-service:latest
```

### Run

##### - Run using Docker-Compose

you can create and start all the services from your configuration (docker-compose.yml) -

```
docker-compose -f docker-compose.yml pull
docker-compose -f docker-compose.yml up
```
