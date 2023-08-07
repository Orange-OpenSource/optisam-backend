@license
Feature: License Service Test - Compliance for Metrics saas : admin

  Background:
  # * def licenseServiceUrl = "https://optisam-license-int.apps.fr01.paas.tech.orange"
    * url licenseServiceUrl+'/api/v1/license'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

  @1
Scenario: Validate Licence for Concurrent Saas metric for product Password_Depots
    Given path 'product', data.concurrent.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200
   # * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.concurrent.SKU+"')].ID")[0]
    
    #* match result.numCptLicences == data.concurrent.Computed_licenses
    #* match result.numAcqLicences == data.concurrent.Acquired_Licenses
    #* match result.deltaNumber == data.concurrent.Delta_licenses
    #* match result.deltaCost == data.concurrent.Delta_Cost
    #* match result.totalCost == data.concurrent.Total_Cost
    #* match result.metric == data.concurrent.metric
    #* match result.computedCost == data.concurrent.computedCost
    #* match result.purchaseCost == data.concurrent.purchaseCost
    #* match result.avgUnitPrice == data.concurrent.avgUnitPrice

    @2
Scenario: Validate Licence for Nominative Saas metric for product FoxEditor
    Given path 'product', data.nominative.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200
   # * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.nominative.SKU+"')].ID")[0]
    
    #* match result.numCptLicences == data.nominative.Computed_licenses
    #* match result.numAcqLicences == data.nominative.Acquired_Licenses
    #* match result.deltaNumber == data.nominative.Delta_licenses
    #* match result.deltaCost == data.nominative.Delta_Cost
    #* match result.totalCost == data.nominative.Total_Cost
    #* match result.metric == data.nominative.metric
    #* match result.computedCost == data.nominative.computedCost
    #* match result.purchaseCost == data.nominative.purchaseCost
    #* match result.avgUnitPrice == data.nominative.avgUnitPrice

    @3
Scenario: Validate Compliance for Concurrent Saas metric for aggregation
    Given path 'aggregation', data.concurrent_agg.name, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200
    #* def result = karate.jsonPath(response, "$.acq_rights[?(@.aggregationName=='"+data.concurrent_agg.name+"')].ID")[0]
    
    #* match result.numCptLicences == data.concurrent_agg.Computed_licenses
    #* match result.numAcqLicences == data.concurrent_agg.Acquired_Licenses
    #* match result.deltaNumber == data.concurrent_agg.Delta_licenses
    #* match result.deltaCost == data.concurrent_agg.Delta_Cost
    #* match result.totalCost == data.concurrent_agg.Total_Cost
    #* match result.metric == data.concurrent_agg.metric
    #* match result.computedCost == data.concurrent_agg.computedCost
    #* match result.purchaseCost == data.concurrent_agg.purchaseCost
    #* match result.avgUnitPrice == data.concurrent_agg.avgUnitPrice 
    
    @4
Scenario: Validate Compliance for Nominative Saas metric for aggregation
        Given path 'aggregation', data.nominative_agg.name, 'acquiredrights'
        And params {scope: '#(scope)'}
        When method get
        Then status 200
       # * def result = karate.jsonPath(response, "$.acq_rights[?(@.aggregationName=='"+data.nominative_agg.name+"')].ID")[0]
       # * match result.numCptLicences == data.nominative_agg.Computed_licenses
       # * match result.numAcqLicences == data.nominative_agg.Acquired_Licenses
       # * match result.deltaNumber == data.nominative_agg.Delta_licenses
       # * match result.deltaCost == data.nominative_agg.Delta_Cost
       # * match result.totalCost == data.nominative_agg.Total_Cost
       # * match result.metric == data.nominative_agg.metric
       # * match result.computedCost == data.nominative_agg.computedCost
       # * match result.purchaseCost == data.nominative_agg.purchaseCost
       # * match result.avgUnitPrice == data.nominative_agg.avgUnitPrice 



