@simulation
Feature: Simulation Service Test : Admin

  Background:
  # * def simulationServiceUrl = "https://optisam-simulation-int.apps.fr01.paas.tech.orange"
    * url simulationServiceUrl+'/api/v1/simulation'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

#  TODO: add simulation for all metrics that exist

 #-------------For Individual----------------------------------------------------#
  

  Scenario: Verify the selection of product of Editor
    Given url 'https://optisam-product-dev.apps.fr01.paas.tech.orange/api/v1/product/editors/products'
    And params { scope:'#(scope)',editor:'#(data.Editor.editor1)'}
    When method get 
    Then status 200
  

  Scenario: To validate the simulation of matric
    Given path 'metric'
    And request data.metricsimulation.request4
    When method post 
    Then status 200
   

    
  Scenario: To validate the matric simulation  with 2 matrics
    Given path 'metric'
    And request data.metricsimulation.request3
    When method post 
    Then status 200 
   


   #---------------------For Aggregation--------------------------------------------#
  

  

  Scenario: Verify the simlation on Aggregation
    Given  path 'metric'
    And request data.metricsimulation.request5
    When method post 
    Then status 200 
  

    #-----------------------------------Cost Simulation----------------------------------------#
  

  Scenario: To verify the cost simulation 
    Given path 'cost'
    And request data.Costsimulation.request
    When method post
    Then status 200
 

  Scenario: To verify the cost simulation for 2 SKU's
    Given path 'cost'
     And request data.Costsimulation.request2
    When method post
    Then status 200  
    


  

 

    #------------------------Hardware-----------------------------------------#

  Scenario: To verify the responce after clicking on Hardware 
    Given url 'https://optisam-equipment-dev.apps.fr01.paas.tech.orange/api/v1/equipment/types'
    And params { scopes:'#(scope)'}
    When method get
    Then status 200
    And match response.equipment_types[*].ID contains ["0x2eaf10","0x2eaf13","0x2eaf1b","0x2eaf28"] 


    # Hardware Simulation is not working   for any of the Equipment Type 


  @hardwaresimulation
  Scenario: verify Hardware Simulation
    Given path 'hardware'
    And request data.hardwaresimulation
    When method post
    Then status 200




