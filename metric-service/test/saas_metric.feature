@metric
Feature: Metric Service Test - Create new SAAS metrics : admin user

  Background:
    #* def metricServiceUrl = "https://optisam-metric-dev.apps.fr01.paas.tech.orange"
    * url metricServiceUrl+'/api/v1'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'DEM'

  
  Scenario: To check if Metric API is working fine
    Given path 'metrics'
    And params {scopes:'#(scope)'}
    When method get
    Then status 200

  Scenario: To create a SAAS nominative standard metric
    Given path 'metric/sns'
    And request data.create_saas_metric_nom
    When method post 
    Then status 200

  Scenario: To edit nominative metric
    Given path 'metric/sns'
    And request data.edit_saas_metric_nom
    When method patch 
    Then status 200

  Scenario: To delete nominative metric
    Given path 'metrics'
    And params {scopes:'#(scope)'}
    When method get
    Then status 200
    And def metric = karate.jsonPath(response, '$.metrices[?(@.type=="saas.nominative.standard")].name')[0]
    * header Authorization = 'Bearer '+access_token
    Given path 'metric/' + metric
    And params {scope:'#(scope)'}
    When method delete
    Then status 200

Scenario: To create a SAAS concurrent standard metric
    Given path 'metric/saas_conc'
    And request data.create_saas_metric_conc
    When method post 
    Then status 200

Scenario: To edit concurrent metric
    Given path 'metric/saas_conc'
    And request data.edit_saas_metric_conc
    When method patch 
    Then status 200

Scenario: To delete concurrent metric
    Given path 'metrics'
    And params {scopes:'#(scope)'}
    When method get
    Then status 200
    #And def metric = karate.jsonPath(response, '$.metrices[?(@.type=="saas.concurrent.standard")].name')[0]
    * header Authorization = 'Bearer '+access_token
    Given path 'metric/' + data.create_saas_metric_conc.Name
    And params {scope:'#(scope)'}
    When method delete
    Then status 200