@license
Feature: License Service Test - Oracle editor metrics : Admin

  Background:
  # * def licenseServiceUrl = "https://optisam-license-int.kermit-noprod-b.itn.intraorange"
    * url licenseServiceUrl+'/api/v1/license'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'


  Scenario: Validate License for ops metric : oracle.processor.standard with base equipment type Server 
    Given path 'product/'+data.ops_server_license.swidTag+'/acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    #  TODO: update license calculation value
    # * match response.acq_rights[0] == data.ops_server_license

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

  Scenario: Validate License for nup metric : oracle.nup.standard with base equipment type Partition 
    Given path 'product/'+data.nup_partition_license.swidTag+'/acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    #  TODO: update license calculation value
    # * match response.acq_rights[0] == data.nup_partition_license

