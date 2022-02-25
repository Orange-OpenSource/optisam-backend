@license
Feature: License Service Test - Compliance for application and aggregation : admin

  Background:
  # * def licenseServiceUrl = "https://optisam-license-int.kermit-noprod-b.itn.intraorange"
    * url licenseServiceUrl+'/api/v1/license'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'


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
