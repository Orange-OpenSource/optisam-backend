@license
Feature: License Service Test - Compliance for application and aggregation : admin

  Background:
  # * def licenseServiceUrl = "https://optisam-license-int.apps.fr01.paas.tech.orange"
    * url licenseServiceUrl+'/api/v1/license'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


  Scenario: Validate Compliance of an application
    Given path 'applications', data.app_license.app_id, 'products',data.app_license.swidTag
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    * remove data.app_license.app_id
    * match response.acq_rights[*] contains data.app_license


  Scenario: Validate Compliance of an Product aggregation
    Given path 'products/aggregations/productview', data.agg_license.aggName, 'acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    # * remove data.agg_license.aggName
    # * match response.acq_rights[*].deltaCost contains data.agg_license.deltaCost
    # * match response.acq_rights[*].numCptLicences contains data.agg_license.numCptLicences
