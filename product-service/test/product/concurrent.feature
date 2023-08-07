Feature: Concurrent User 

Background:
# * def productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
  * url productServiceUrl+'/api/v1/product'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('../common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'

Scenario: To export Concurrent User --Individual 
    Given path 'concurrent/users/export'
    And params {sort_by:'aggregation_name',sort_order:'asc',scopes:'#(scope)',is_aggregation:'false'}
    When method get 
    Then status 200 

Scenario: To export Concurrent User --Aggregation
    Given path 'concurrent/users/export'
    And params {sort_by:'aggregation_name',sort_order:'asc',scopes:'#(scope)',is_aggregation:'true'}
    When method get 
    Then status 200 

    #----------------History of concurrent users---------------------------#
Scenario: To view concurrent Users --Individual
  Given path 'concurrent'
  And params {page_num: 1,page_size: 50,sort_by: 'purchase_date',sort_order: 'asc',scopes: '#(scope)',is_aggregation: 'false',search_params.product_editor.filteringkey:'#(data.Concurrent_History.editor)',search_params.product_editor.filter_type: 'true',search_params.product_name.filteringkey: '#(data.Concurrent_History.product)',search_params.product_name.filter_type: 'true',search_params.product_version.filteringkey: '#(data.Concurrent_History.version)',search_params.product_version.filter_type: 'true'}
  When method get 
  Then status 200 

Scenario: To view the History of Concurrent User-Individual
  Given path 'concurrent',scope
  And params {swidtag:'#(data.Concurrent_History.swidtag)',start_date: '#(data.Concurrent_History.start_date)',end_date: '#(data.Concurrent_History.end_date)'}
  When method get 
  Then status 200
  
Scenario: To modify the date to view the history of Concurrent User
  Given path 'concurrent',scope
  And params {swidtag:'#(data.Concurrent_History.swidtag)',start_date: '#(data.Concurrent_History.start_date1)',end_date: '#(data.Concurrent_History.end_date1)'}
When method get 
Then status 200

Scenario: To verify searching on Concurrent User-Individual
  Given path 'concurrent'
  And params { page_num: 1,page_size: 50,sort_by: 'purchase_date',sort_order: 'asc',scopes: '#(scope)',is_aggregation: 'false',search_params.product_editor.filteringkey: '#(data.Concurrent_History.editor)',search_params.product_editor.filter_type: 'true',search_params.product_name.filteringkey: '#(data.Concurrent_History.product)', search_params.product_name.filter_type: 'true', search_params.product_version.filteringkey:, search_params.product_version.filter_type: 'true', search_params.profile_user.filteringkey: '#(data.Concurrent_History.product)'}
  When method get 
  Then status 200 

Scenario Outline: To verify the pagination on concurrent User -Individual
  Given path 'concurrent'
  And params {page_num:1,page_size: <page_size>,sort_by: 'purchase_date',sort_order: 'asc',scopes: '#(scope)',is_aggregation: 'false',search_params.product_editor.filteringkey:'#(data.Concurrent_History.editor)',search_params.product_editor.filter_type: 'true',search_params.product_name.filteringkey: '#(data.Concurrent_History.product)',search_params.product_name.filter_type: 'true',search_params.product_version.filteringkey: '#(data.Concurrent_History.version)',search_params.product_version.filter_type: 'true'}
  When method get 
  Then status 200
Examples:
|page_size|
|50|
|100|
|200|

Scenario Outline: To verify the pagination on concurrent User -Individual(With Invalid Inputs)
 Given path 'concurrent'
  And params {page_num:1,page_size: <page_size>,sort_by: 'purchase_date',sort_order: 'asc',scopes: '#(scope)',is_aggregation: 'false',search_params.product_editor.filteringkey:'#(data.Concurrent_History.editor)',search_params.product_editor.filter_type: 'true',search_params.product_name.filteringkey: '#(data.Concurrent_History.product)',search_params.product_name.filter_type: 'true',search_params.product_version.filteringkey: '#(data.Concurrent_History.version)',search_params.product_version.filter_type: 'true'}
  When method get 
  Then status 400
Examples:
|page_size|
|"asd"|
|"A"|
|20|
|19|
|205|
|1|

#------------------Aggregation_---------------#
Scenario: To view the concurrent Users for Aggregations
  Given path 'concurrent'
  And params { page_num: 1,page_size: 50,sort_by: purchase_date,sort_order: asc,scopes: '#(scope)',is_aggregation: true,search_params.aggregation_name.filteringkey:'#(data.Concurrent_agg_History.aggregation_name)',search_params.aggregation_name.filter_type: true,}
  When method get 
  Then status 200 

Scenario: To view the Concurrent History for Aggregations
  Given path 'concurrent',scope
And params {aggID: '#(data.Concurrent_agg_History.aggregation_id)',start_date:'#(data.Concurrent_agg_History.start_date)',end_date: '#(data.Concurrent_agg_History.end_date)'}
When method get 
Then status 200 
Scenario Outline: To verify searching and sorting on the Concurrent Aggregation user (positive)
  Given path 'concurrent'
  And params { page_num: 1,page_size: 50,sort_by: purchase_date,sort_order: 'asc',scopes: '#(scope)',is_aggregation: 'true',search_params.aggregation_name.filteringkey: '#(data.Concurrent_agg_History.aggregation_name)',search_params.aggregation_name.filter_type: 'true',search_params.team.filteringkey: '#(data.Concurrent_agg_History.team)',search_params.profile_user.filteringkey: '#(data.Concurrent_agg_History.profile_user)',search_params.number_of_users.filteringkey: '#(data.Concurrent_agg_History.number_of_users)',}
  When method get 
  Then status 200 
  Examples:
  |page_no.|
  |50|
  |100|
  |200|

Scenario Outline: To verify searching and sorting on the Concurrent Aggregation user (Negative)
  Given path 'concurrent'
  And params { page_num: 1,page_size: 50,sort_by: purchase_date,sort_order: 'asc',scopes: '#(scope)',is_aggregation: 'true',search_params.aggregation_name.filteringkey: '#(data.Concurrent_agg_History.aggregation_name)',search_params.aggregation_name.filter_type: 'true',search_params.team.filteringkey: '#(data.Concurrent_agg_History.team)',search_params.profile_user.filteringkey: '#(data.Concurrent_agg_History.profile_user)',search_params.number_of_users.filteringkey: '#(data.Concurrent_agg_History.number_of_users)',}
  When method get 
  Then status 200 
  Examples:
  |page_no.|
  |"a"|
  |"A"|
  |11|
