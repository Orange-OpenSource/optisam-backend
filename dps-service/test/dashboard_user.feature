@dashboard
Feature: Dashboard Test

  Background:
    * url dpsServiceUrl+'/api/v1/dps'
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


  @get
  Scenario: Developement Rate
    Given path 'dashboard/quality'
    * params {noOfDataPoints:6, frequency: 'MONTHLY',scope:'#(scope)'}
    When method get
    Then status 200
    # And match response contains data.quality
