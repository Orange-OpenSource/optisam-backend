@dashboard
Feature: Dashboard Test on dps: admin user

  Background:
    * url dpsServiceUrl+'/api/v1/dps'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
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
    
  @SmokeTest
    @schema
   Scenario: Schema validation for Developement Rate on Quality dashboard
   Given path 'dashboard/quality'
    * params {noOfDataPoints:6, frequency: 'MONTHLY',scope:'#(scope)'}
    * def schema = data.schema_quality
    When method get
    Then status 200


 

  


    