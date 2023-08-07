Feature: Acquired Rights Shared License Service Testing --Admin

Background:
# * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
  * url productServiceUrl+'/api/v1/product'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('../common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'
 
  
@SmokeTest
  Scenario: Fetch Acquired Rights
    Given path 'acqrights'
    And  params { page_num:1, page_size:20, sort_by:'SKU', sort_order:'desc', scopes:'#(scope)'}
    When method get 
    And print response
    Then status 200
    

  Scenario: Search AcquiredRights by multiple  column
    Given path 'acqrights'
    And params { page_num:1, page_size:50, sort_by:'SKU', sort_order:'asc', scopes:'#(scope)'}
    And params {search_params.swidTag.filteringkey: '#(data.getAcqrights.swid_tag)'}
    And params {search_params.productName.filteringkey: '#(data.getAcqrights.product_name)'}
    When method get
    Then status 200
    And response.totalRecords > 0
    And match   response.acquired_rights[*].swid_tag contains data.getAcqrights.swid_tag
    And match each response.acquired_rights[*].product_name == data.getAcqrights.product_name
   # And match  response.acquired_rights contains data.getAcqrights

  Scenario: To get the nominative users
    Given path 'nominative/users'
    And params { page_num:1,page_size:50,sort_by:'activation_date',sort_order:'asc',scopes:'API' }
    When method get 
    Then status 200 
#--------------------------------Shared Licenses-----------------------------------------------#
  Scenario: To get the details about license status for Acquired Right
    Given path 'acqrights/licenses'
    And params {scope:'#(scope)',sku: '#(data.AcqWithSharedLic.SKU)'}
    When method get 
    Then status 200 
    #* match response.available_licenses == data.AcqWithSharedLic.available_licenses
  
  Scenario: To verfiy Sharing of  License with  other scope
    Given path 'licenses'
    And request data.Sharing_LIC
    When method put
    Then status 200 
    And match response.success == true 

  Scenario Outline: To verify the sharing the licenses by entering Invalid Inputs
    Given path 'licenses'
    * set data.Sharing_LIC.license_data[*].shared_licenses = <Number>
   And request data.Sharing_LIC
  When method put
  Then status 400 
  Examples:
  |Number|
  |-100|
  |4.5|
  |"twenty four"|
 
  

  

Scenario: To verfiy Whether Licenses Can be shared with  more then one scope
  Given path 'licenses'
  And request data.Sharing_LIC_Mul
  When method put
  Then status 200 
  And match response.success == true 


Scenario: To verfiy Whether Licenses Can be shared by Keeping Source and Destination same 
  Given path 'licenses'
  And request data.Sharing_LIC_Same
  When method put
  Then status 200 
  And match response.success == true 


Scenario: To varify  whether i can share Licenses more Than Available licenses 
  Given path 'licenses'
  And request data.Sharing_LIC_Mul_Max
  When method put
  Then status 400 
  And match response.message == "LicencesNotAvailable"

Scenario: To verify sharing lesser number of licenses then previously shared 
  Given path 'licenses'
  * set data.Sharing_LIC.license_data[*].shared_licenses = 1
  And request data.Sharing_LIC
  When method put
  Then status 200 
  And match response.success == true



Scenario: To verify sharing license to same scope multiple times 
  Given path 'licenses'
  And request data.Sharing_LIC_Mul_SameScope
  When method put
  Then status 200 
  And match response.success == true

Scenario: To get the status of License of Aggregated Acquired Rights 
  Given path 'aggregated_acqrights'
  And params {scope:'#(scope)',page_num:1,page_size:50,sort_by:"SKU",sort_order:"asc"}
  When method get 
  Then status 200
 #* match response.aggregations.swidtags[*] == data.AgAcqAcqWithSharedLic.swidtags

Scenario: To verify sharing of Aggregated Acquired Rights with other scope 
  Given path 'aggrights/licenses'
  And request data.Sharing_Agg_Lic
  When method put 
  Then status 200
  And match response.success == true


Scenario Outline: To verify sharing of Aggregated Acquired Rights with Invalid Input
  Given path 'aggrights/licenses'
  * set data.Sharing_Agg_Lic.license_data[*].shared_licenses = <Number>
  And request data.Sharing_Agg_Lic
 When method put
 Then status 400 
 Examples:
 |Number|
 |-100|
 |4.5|
 |"twenty four"|


Scenario: To verify sharing of Aggregated Acquired Rights with other scope 
  Given path 'aggrights/licenses'
  * set data.Sharing_Agg_Lic.license_data[*].shared_licenses = 0
  And request data.Sharing_Agg_Lic
  When method put 
  Then status 200
  And match response.success == true

Scenario: To verify sharing license to same scope multiple times 
  Given path 'aggrights/licenses'
  And request data.Sharing_Agg_Lic_Same_Mul
  When method put
  Then status 200
  
  








  



































    








