@dashboard
Feature: Dashboard Test

  Background:
    * url dpsServiceUrl+'/api/v1/dps'
  # * def dpsServiceUrl = "https://optisam-dps-int.kermit-noprod-b.itn.intraorange"
    * def credentials = {username:'testuser@test.com', password: 'password'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'


  @get
  Scenario: Developement Rate
    Given path 'dashboard/quality'
    * params {noOfDataPoints:6, frequency: 'MONTHLY',scope:'#(scope)'}
    When method get
    Then status 200
    # And match response contains data.quality


  # @get @ignore
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

  @get @ignore

  Scenario: Data Failure Rate when there is no data in scope
    Given path 'dashboard/quality/datafailurerate'
    * params {scope:'CLR'}
    When method get
    Then status 404

  @get @ignore
  
  Scenario: Data Failure Ratio when there is no data in scope
    Given path 'dashboard/quality/failurereasonsratio'
    * params {scope:'CLR'}
    When method get
    Then status 404