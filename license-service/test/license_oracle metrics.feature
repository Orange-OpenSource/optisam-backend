@license
Feature: License Service Test - Oracle editor metrics : Admin

  Background:
  # * def licenseServiceUrl = "https://optisam-license-int.apps.fr01.paas.tech.orange"
    * url licenseServiceUrl+'/api/v1/license'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


  Scenario: Validate License for ops metric : oracle.processor.standard with base equipment type Server 
    Given path 'product/'+data.ops_server_license.swidTag+'/acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200


  Scenario: Validate Compliance of an Product aggregation for Oracle_RAC
    Given path 'aggregation', data.aggregationName_Oracle_RAC.aggregationName , 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.aggregationName_Oracle_RAC.SKU+"')]")[0]

    * match result.numCptLicences == data.aggregationName_Oracle_RAC.Computed_licenses
    * match result.numAcqLicences == data.aggregationName_Oracle_RAC.Acquired_Licenses
    * match result.deltaNumber == data.aggregationName_Oracle_RAC.Delta_licenses


  Scenario: Validate Compliance of an Product aggregation for Oracle_enterprise_database
    Given path 'aggregation', data.aggregationName_Oracle_enterprise_database.aggregationName , 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.aggregationName_Oracle_enterprise_database.SKU+"')]")[0]

    * match result.numCptLicences == data.aggregationName_Oracle_enterprise_database.Computed_licenses
    * match result.numAcqLicences == data.aggregationName_Oracle_enterprise_database.Acquired_Licenses

  Scenario: Validate Compliance of an Product aggregation for oracle_learning_management
    Given path 'aggregation', data.oracle_learning_management.SKU , 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.oracle_learning_management.SKU+"')]")[0]

    * match result.numCptLicences == data.oracle_learning_management.Computed_licenses
    * match result.numAcqLicences == data.oracle_learning_management.Acquired_Licenses

  Scenario: Validate Licence for oracle.processor for product Oracle Weblogic Server
    Given path 'product', data.Oracle_Weblogic_Server_Oracle_Weblogic.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200
  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.Oracle_Weblogic_Server_Oracle_Weblogic.SKU+"')]")[0]
    

    * match result.numCptLicences == data.Oracle_Weblogic_Server_Oracle_Weblogic.Computed_licenses
    * match result.numAcqLicences == data.Oracle_Weblogic_Server_Oracle_Weblogic.Acquired_Licenses
    * match result.deltaNumber == data.Oracle_Weblogic_Server_Oracle_Weblogic.Delta_licenses
    * match result.deltaCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.Delta_Cost
    * match result.totalCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.Total_Cost
    * match result.metric == data.Oracle_Weblogic_Server_Oracle_Weblogic.metric
    * match result.computedCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.computedCost
    * match result.purchaseCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.purchaseCost
    * match result.avgUnitPrice == data.Oracle_Weblogic_Server_Oracle_Weblogic.avgUnitPrice
