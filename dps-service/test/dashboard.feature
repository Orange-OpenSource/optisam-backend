@dashboard
Feature: Dashboard Test on dps: admin user

  Background:
    * url dpsServiceUrl+'/api/v1/dps'
  # * def dpsServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API' 

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


 

  


    