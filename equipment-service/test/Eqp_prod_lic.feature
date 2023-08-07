Feature: Equipment Product License Test 

Background:
 #* def productServiceUrl = "https://optisam-license-dev.apps.fr01.paas.tech.orange"
  * url licenseServiceUrl+'/api/v1/license/'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'
  
  #----------------  Compliance of product for server----------------------------#
@complience
Scenario: To Get the detail and Complience  of Product of server
    Given path 'product' ,data.server.product_swidTag,'acquiredrights'
    And params {scope:'#(scope)' }
    When method get 
    Then status 200
   # And match response.acq_rights[*].productName contains ["Adobe Media Server"]

 #----------------  Compliance of product for Softpartition----------------------------#
@complience
Scenario: To get the detail and complience  of product of Softpartition
    Given path 'product' ,data.Softpartition.product_swidTag,'acquiredrights'
    And params {scope:'#(scope)' }
    When method get 
    Then status 200
   # And match response.acq_rights[*].productName contains ["Adobe Media Server"]





  
