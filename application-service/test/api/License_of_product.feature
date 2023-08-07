Feature: To get the details of Product 

Background:
  * url licenseServiceUrl+'/api/v1'
  #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'


  Scenario: To get the details of product 
    Given path 'license/applications', data.product.application_id ,'products',data.product.product_name
    And params { scope:'#(scope)'}
    When method get 
    Then status 200
    
