Feature: Nominative User 

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
Scenario: To export Nominative Users - For individual 
    Given path 'nominative/users/export'
    And params {sort_by:'aggregation_name',sort_order:'asc',scopes:'#(scope)',is_product:'true'}
    When method get 
    Then status 200 
    
  @SmokeTest
Scenario: To export Nominative Users - For Aggregation
    Given path 'nominative/users/export'
    And params {sort_by:'aggregation_name',sort_order:'asc',scopes:'#(scope)',is_product:'false'}
    When method get 
    Then status 200 

  #---------------Nominative Users History----------------#
Scenario: To View the nominative User-- Individual 
  Given path 'nominative/users'
  And params { page_num: 1, page_size: 50,sort_by:'activation_date',sort_order: 'asc',scopes: '#(scope)',is_product: 'true',search_params.editor.filteringkey:'#(data.nominative_History.editor)',search_params.editor.filter_type: 'true', search_params.product_name.filteringkey: '#(data.nominative_History.product_name)', search_params.product_name.filter_type: 'true', search_params.product_version.filteringkey: '#(data.nominative_History.product_version)', search_params.product_version.filter_type: 'true'}
   When method get 
   Then status 200 
  #And match response.nominative_user[*].product_name == ["Irisa"]

Scenario Outline: To verfiy the Searching  and sorting on Nominative -Individual(Positive)
  Given path 'nominative/users'
  And params {page_num: 1,page_size: 50,sort_by: activation_date,sort_order: 'asc',scopes: '#(scope)',is_product: true,search_params.editor.filteringkey: '#(data.nominative_History.editor)', search_params.editor.filter_type: 'true',search_params.product_name.filteringkey: '#(data.nominative_History.product_name)', search_params.product_name.filter_type: 'true', search_params.product_version.filteringkey: '#(data.nominative_History.product_version)', search_params.product_version.filter_type: 'true', search_params.first_name.filteringkey: '#(data.nominative_History.first_name)', search_params.user_email.filteringkey: '#(data.nominative_History.user_email)', search_params.user_name.filteringkey: '#(data.nominative_History.user_name)'}
  When method get 
  Then status 200
Examples:
|page_size|
|50|
|100|
|200|

Scenario Outline: To verfiy the Searching  and sorting on Nominative -Individual(Negative)
  Given path 'nominative/users'
  And params {page_num: 1,page_size:<page_size>,sort_by: activation_date,sort_order: 'asc',scopes: '#(scope)',is_product: true,search_params.editor.filteringkey: '#(data.nominative_History.editor)', search_params.editor.filter_type: 'true',search_params.product_name.filteringkey: '#(data.nominative_History.product_name)', search_params.product_name.filter_type: 'true', search_params.product_version.filteringkey: '#(data.nominative_History.product_version)', search_params.product_version.filter_type: 'true', search_params.first_name.filteringkey: '#(data.nominative_History.first_name)', search_params.user_email.filteringkey: '#(data.nominative_History.user_email)', search_params.user_name.filteringkey: '#(data.nominative_History.user_name)'}
  When method get 
  Then status 400
Examples:
|page_size|
|"a"|
|"A"|
|19|


#-------Nominative Aggregation Users-----------#
Scenario: To get the individual Aggregation User
  Given path 'nominative/users'
  And params {page_num: 1,page_size: 50,sort_by: 'activation_date',sort_order: 'asc',scopes: '#(scope)',is_product: 'false',search_params.aggregation_name.filteringkey:'#(data.nominative_agg_History.aggregation_name)',search_params.aggregation_name.filter_type: 'true'}
When method get 
Then status 200 

Scenario Outline: To searching and sorting on the Individual Aggregation-Positive
  Given path 'nominative/users'
  And params {page_num: 1, page_size: 50, sort_by: 'activation_date', sort_order: 'asc', scopes: '#(scope)', is_product: 'false', search_params.aggregation_name.filteringkey:'#(data.nominative_agg_History.aggregation_name)', search_params.aggregation_name.filter_type: 'true', search_params.first_name.filteringkey: '#(data.nominative_agg_History.first_name)', search_params.profile.filteringkey:'#(data.nominative_agg_History.profile)', search_params.user_email.filteringkey:'#(data.nominative_agg_History.user_email)', search_params.user_name.filteringkey:'#(data.nominative_agg_History.user_name)'}
  When method get 
  Then status 200
  Examples:
  |page_size|
  |50|
  |100|
  |200|

Scenario Outline: To searching and sorting on the Individual Aggregation--Negative
  Given path 'nominative/users'
  And params {page_num: 1, page_size: 50, sort_by: 'activation_date', sort_order: 'asc', scopes: '#(scope)', is_product: 'false', search_params.aggregation_name.filteringkey:'#(data.nominative_agg_History.aggregation_name)', search_params.aggregation_name.filter_type: 'true', search_params.first_name.filteringkey: '#(data.nominative_agg_History.first_name)', search_params.profile.filteringkey:'#(data.nominative_agg_History.profile)', search_params.user_email.filteringkey:'#(data.nominative_agg_History.user_email)', search_params.user_name.filteringkey:'#(data.nominative_agg_History.user_name)'}
  When method get 
  Then status 200
  Examples:
  |page_size|
  |"a"|
  |"A"|
  |19|
