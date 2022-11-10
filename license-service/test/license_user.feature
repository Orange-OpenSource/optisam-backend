@license
Feature: License Service Test

  Background:
  # * def licenseServiceUrl = "https://optisam-license-int.apps.fr01.paas.tech.orange"
    * url licenseServiceUrl+'/api/v1/license'
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

# Oracle
  Scenario: Validate License for ops metric : oracle.processor.standard with base equipment type Partition 
    Given path 'product/'+data.ops_partition_license.swidTag+'/acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    * match response.acq_rights[0] == data.ops_partition_license

  Scenario: Validate License for nup metric : oracle.nup.standard with base equipment type Server 
    Given path 'product/'+data.nup_server_license.swidTag+'/acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    * match response.acq_rights[0] == data.nup_server_license

  Scenario: Validate License for sag metric : sag
    Given path 'product/'+data.sag_license.swidTag+'/acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
     * match response.acq_rights[0] == data.sag_license

  Scenario: Validate License for inm metric : os_instance
    Given path 'product/'+data.inm_license.swidTag+'/acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
     * match response.acq_rights[0] == data.inm_license


  Scenario: Validate License for pvu metric : ibm_pvu
    Given path 'product/'+data.ibm_pvu_license.swidTag+'/acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
     * match response.acq_rights[0] == data.ibm_pvu_license


## TODO : update license value
  Scenario: Validate License for acs metric : attribute_counter_core
    Given path 'product/'+data.acs_license.swidTag+'/acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
     * match response.acq_rights[0] == data.acs_license


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
    * remove data.agg_license.aggName
    * match response.acq_rights[*].deltaCost contains data.agg_license.deltaCost
    * match response.acq_rights[*].numCptLicences contains data.agg_license.numCptLicences
