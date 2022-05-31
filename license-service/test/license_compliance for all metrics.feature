@license
Feature: License Service Test - Compliance for Metrics inm,acs,sag,pvu : admin

  Background:
  # * def licenseServiceUrl = "https://optisam-license-int.apps.fr01.paas.tech.orange"
    * url licenseServiceUrl+'/api/v1/license'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'


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
