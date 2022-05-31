@dashboard
Feature: Dashboard Test on dps: admin user

  Background:
    * url dpsServiceUrl+'/api/v1/dps'
  # * def dpsServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT' 

  # TODO: update dashboard tests for cron handling 
  @get
  Scenario: Developement Rate
    Given path 'dashboard/quality'
    * params {noOfDataPoints:6, frequency: 'MONTHLY',scope:'#(scope)'}
    When method get
    Then status 200
    # And match response contains data.quality

    @schema
   Scenario: Schema validation for Developement Rate on Quality dashboard
   Given path 'dashboard/quality'
    * params {noOfDataPoints:6, frequency: 'MONTHLY',scope:'#(scope)'}
    * def schema = data.schema_quality
    When method get
    Then status 200


  @get  @ignore
  
  # Scenario: Data Failure Rate
  #   Given path 'dashboard/quality/datafailurerate'
  #   * params {scope:'#(scope)'}
  #   When method get
  #   Then status 200

  # @get
  # Scenario: Data Failure Ratio
  #   Given path 'dashboard/quality/failurereasonsratio'
  #   * params {scope:'#(scope)'}
  #   When method get 
  #   * call pause 2000
  #   Then status 200 
    


# Handling of new scope - CLR for different environments

  @ignore @dev

  Scenario: Data Failure Rate when there is no data in scope
    Given path 'dashboard/quality/datafailurerate'
    * params {scope:'CLR'}
    When method get
    Then status 404

  @ignore @dev

  Scenario: Data Failure Ratio when there is no data in scope
    Given path 'dashboard/quality/failurereasonsratio'
    * params {scope:'CLR'}
    When method get
    Then status 404


 

  


    