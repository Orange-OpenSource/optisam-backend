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
    Given path 'product', data.product.swidTag, 'acquiredrights'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
   # * match response.acq_rights[*].swidTag contains ["Adobe_Media_Server_Adobe_5.0.16"]


  Scenario: Validate Compliance of an Product aggregation for oracle_learning_management
    Given path 'aggregation', data.oracle_learning_management.SKU , 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.oracle_learning_management.SKU+"')]")[0]

    * match result.numCptLicences == data.oracle_learning_management.Computed_licenses
    * match result.numAcqLicences == data.oracle_learning_management.Acquired_Licenses
    * match result.deltaNumber == data.oracle_learning_management.Delta_licenses
    * match result.deltaCost == data.oracle_learning_management.Delta_Cost
    * match result.totalCost == data.oracle_learning_management.Total_Cost
    * match result.metric == data.oracle_learning_management.metric
    * match result.computedCost == data.oracle_learning_management.computedCost
    * match result.purchaseCost == data.oracle_learning_management.purchaseCost
    * match result.avgUnitPrice == data.oracle_learning_management.avgUnitPrice

  Scenario: Validate Compliance of an Product aggregation for redhat_openshift_standard
    Given path 'aggregation', data.aggregationName_redhat_openshift.aggregationName , 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.aggregationName_redhat_openshift.SKU+"')]")[0]

    * match result.numCptLicences == data.aggregationName_redhat_openshift.Computed_licenses
    * match result.numAcqLicences == data.aggregationName_redhat_openshift.Acquired_Licenses
    * match result.deltaNumber == data.aggregationName_redhat_openshift.Delta_licenses
    * match result.deltaCost == data.aggregationName_redhat_openshift.Delta_Cost
    * match result.totalCost == data.aggregationName_redhat_openshift.Total_Cost
    * match result.metric == data.aggregationName_redhat_openshift.metric
    * match result.computedCost == data.aggregationName_redhat_openshift.computedCost
    * match result.purchaseCost == data.aggregationName_redhat_openshift.purchaseCost
    * match result.avgUnitPrice == data.aggregationName_redhat_openshift.avgUnitPrice 

  Scenario: Validate Compliance of an Product aggregation for redhat_CS
    Given path 'aggregation', data.aggregationName_redhat_CS.aggregationName , 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.aggregationName_redhat_CS.SKU+"')]")[0]

    * match result.numCptLicences == data.aggregationName_redhat_CS.Computed_licenses
    * match result.numAcqLicences == data.aggregationName_redhat_CS.Acquired_Licenses
    * match result.deltaNumber == data.aggregationName_redhat_CS.Delta_licenses
    * match result.deltaCost == data.aggregationName_redhat_CS.Delta_Cost
    * match result.totalCost == data.aggregationName_redhat_CS.Total_Cost
    * match result.metric == data.aggregationName_redhat_CS.metric
    * match result.computedCost == data.aggregationName_redhat_CS.computedCost
    * match result.purchaseCost == data.aggregationName_redhat_CS.purchaseCost
    * match result.avgUnitPrice == data.aggregationName_redhat_CS.avgUnitPrice 


  Scenario: Validate Compliance of an Product aggregation for Oracle_RAC
    Given path 'aggregation', data.aggregationName_Oracle_RAC.aggregationName , 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.aggregationName_Oracle_RAC.SKU+"')]")[0]

    * match result.numCptLicences == data.aggregationName_Oracle_RAC.Computed_licenses
    * match result.numAcqLicences == data.aggregationName_Oracle_RAC.Acquired_Licenses
    * match result.deltaNumber == data.aggregationName_Oracle_RAC.Delta_licenses
    * match result.deltaCost == data.aggregationName_Oracle_RAC.Delta_Cost
    * match result.totalCost == data.aggregationName_Oracle_RAC.Total_Cost
    * match result.metric == data.aggregationName_Oracle_RAC.metric
    * match result.computedCost == data.aggregationName_Oracle_RAC.computedCost
    * match result.purchaseCost == data.aggregationName_Oracle_RAC.purchaseCost
    * match result.avgUnitPrice == data.aggregationName_Oracle_RAC.avgUnitPrice


  Scenario: Validate Compliance of an Product aggregation for Oracle_enterprise_database
    Given path 'aggregation', data.aggregationName_Oracle_enterprise_database.aggregationName , 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.aggregationName_Oracle_enterprise_database.SKU+"')]")[0]

    * match result.numCptLicences == data.aggregationName_Oracle_enterprise_database.Computed_licenses
    * match result.numAcqLicences == data.aggregationName_Oracle_enterprise_database.Acquired_Licenses
    * match result.deltaNumber == data.aggregationName_Oracle_enterprise_database.Delta_licenses
    * match result.deltaCost == data.aggregationName_Oracle_enterprise_database.Delta_Cost
    * match result.totalCost == data.aggregationName_Oracle_enterprise_database.Total_Cost
    * match result.metric == data.aggregationName_Oracle_enterprise_database.metric
    * match result.computedCost == data.aggregationName_Oracle_enterprise_database.computedCost
    * match result.purchaseCost == data.aggregationName_Oracle_enterprise_database.purchaseCost
    * match result.avgUnitPrice == data.aggregationName_Oracle_enterprise_database.avgUnitPrice