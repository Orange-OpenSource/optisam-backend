Feature: Equipment Product Service Test

Background:
 # * def productServiceUrl = "https://optisam-product-dev.apps.fr01.paas.tech.orange"
  * url productServiceUrl+'/api/v1'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'
@get
Scenario: Get Details of Product of Equipment -Softpartition
    Given  path 'products'
    And params {page_num:1,page_size:50,sort_by:'swidtag',sort_order:'asc',scopes:'#(scope)', search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.Softpartition.ID)' }
    When method get
    Then status 200
    And match response.totalRecords == 4
    
   
@get
Scenario Outline: Get Details of Product of Equipment (Sorting by name) -Softpartition
    Given  path 'products'
    And params {page_num:1,page_size:50,sort_by:<sortBy>,sort_order:<sortOrder>,scopes:'#(scope)', search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.Softpartition.ID)' }
    When method get
    Then status 200
    Examples:
      | sortBy | sortOrder |  
      | name | asc |
      | name| desc | 
@pagination
Scenario Outline:To verify Pagination and sorting on  Product of Equipemt -softpartition
    Given  path 'products'
    And params {page_num:1,page_size:<page_size>,sort_by:<sortBy>,sort_order:<sortOrder>,scopes:'#(scope)', search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.Softpartition.ID)' }
    When method get
    Then status 200
    Examples:
    | page_size | sortBy|sortOrder|
    | 50        | name  |  asc    |
    | 100       |name   |  desc   |
    | 200       |swidtag|   asc   |
    | 100       |editor |   asc   |

@search
#Scenario: To verify searching on the Product of Equipment with one input in Advance search-softpartition
#    Given   path 'products'
#    And params {page_num:1,page_size:50,sort_by:'swidtag',sort_order:'asc',scopes:'#(scope)', search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.Softpartition.ID)',search_params.editor.filteringkey:'#(data.Softpartition.product_Editor)' }
#    When method get
#    Then status 200
#    And match response.products[*].totalCost == [20200]


    




@search
Scenario Outline: To verify searching on the Product of Equipment with double input  in Advance search-softpartition
    Given  path 'products'
    And params { page_num:1,page_size:100,sort_by:'name',sort_order:'desc',scopes:'#(scope)', search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.Softpartition.ID)',search_params.swidTag.filteringkey:<SwidTag>,search_params.editor.filteringkey:<editor>}
    When method get
    Then status 200
    Examples:
    |                    SwidTag           |          editor                       |
    | #(data.Softpartition.product_swidTag)| #(data.Softpartition.product_Editor)  |  
    |#(data.Softpartition.product_swidTag) |                ""                     |
    |         ""                           |                ""                     |
    |         ""                           | #(data.Softpartition.product_Editor)  |
    |      Invalid Input                   |         invalid input                 |

   

@search
Scenario: To verify searching on the the product with wrong input in advance search -Softpartition
    Given  path 'products'
    And params { page_num:1,page_size:100,sort_by:'name',sort_order:'desc',scopes:'#(scope)', search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.Softpartition.ID)',search_params.swidTag.filteringkey:'#(data.Softpartition.product_swidTag)',search_params.editor.filteringkey:'#(data.Softpartition.product_Editor)'}
    When method get
    Then status 200












#---------------------------------Server-----------------------------------#
@get
Scenario: To get the product of Equipment -Server
    Given path 'products'
    And params {page_num:1,page_size:50,sort_by:'swidtag',sort_order:'asc',scopes:'#(scope)',search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.server.server_id)'}
    When method get 
    Then status 200

@get
Scenario Outline: To  check sorting and pagination on product of Equipment -server 
Given  path 'products'
And params { page_num:1,page_size:<page_size>,sort_by:<sortBy>,sort_order:<sortOrder>,scopes:'#(scope)',search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.server.server_id)',search_params.swidTag.filteringkey:'#(data.Softpartition.product_Editor)' }
When method get 
Then status 200

Examples:
    | page_size | sortBy| sortOrder |
    | 50        | name  |    asc|
    | 100       |name   | desc|
    | 200       |swidtag|   asc|
    |  50       | name  |desc|
@search
Scenario Outline: To  check sorting and pagination on product of Equipment -server 
    Given  path 'products'
    And params { page_num:1,page_size:<page_size>,sort_by:name,sort_order:<sortOrder>,scopes:'#(scope)',search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.server.server_id)',search_params.swidTag.filteringkey:'#(data.Softpartition.product_Editor)' }
    When method get 
    Then status 400
    
    Examples:
        | page_size | sortOrder |
        | 5       |    asc|
        | 1      | desc|
        | 2       |   asc|
        |  "A"      |desc|

    @search
Scenario Outline: To  check searching on product of Equipment -server 
    Given path 'products'
    And params { page_num:1,page_size:50,sort_by:<sortBy>,sort_order:'asc',scopes:'#(scope)',search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.server.server_id)',search_params.swidTag.filteringkey:<SwidTag>,search_params.editor.filteringkey:<editor> }
    When method get 
    Then status 200
    
    Examples:
        | sortBy|                   SwidTag            |          editor                     |
        | name  | #(data.Softpartition.product_swidTag)| #(data.Softpartition.product_Editor)|  
        |name   |#(data.Softpartition.product_swidTag) |                ""                   |
        |swidtag|       ""                             |                ""                   |
        |editor |        ""                            | #(data.Softpartition.product_Editor)|
        | name  | Invalid Input                        |         invalid input               |
        
@search
Scenario: To check the searching on the Product of server with single input
    Given path 'products'
    And params {page_num:1,page_size:50,sort_by:'swidtag',sort_order:'asc',scopes:'#(scope)',search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'server_aixa',search_params.name.filteringkey:'#(data.server.product_name)',search_params.editor.filteringkey:'#(data.server.editor_name)'}
    When method get 
    Then status 200
@search  
Scenario: To check the searching on the Product of server with double input
    Given path 'products'
    And params {page_num:1,page_size:50,sort_by:'swidtag',sort_order:'asc',scopes:'#(scope)',search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'server_aixa',search_params.name.filteringkey:'#(data.server.product_name)',search_params.editor.filteringkey:'#(data.server.editor_name)'}
    When method get 
    Then status 200
@search
Scenario: To check the searching on the Product of server with Invalid input
    Given path 'products'
    And params {page_num:1,page_size:50,sort_by:'swidtag',sort_order:'asc',scopes:'#(scope)',search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'server_aixa',search_params.name.filteringkey:'Invalid ',search_params.editor.filteringkey:'#(data.server.editor_name)'}
    When method get 
    Then status 200

#-------------------------------------------Cluster----------------------------------- ---#
@get
Scenario: To get the details of product of cluster
Given path 'products'
And params {page_num:1,page_size:50,sort_by:'swidtag',sort_order:'asc',scopes:'#(scope)',search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.Cluster.cluster_name)'}
When method get 
Then status 200

#--------------------------------Vcenter--------------------------------------------------#
@get
Scenario: To get the details of product of cluster
    Given path 'products'
    And params {page_num:1,page_size:50,sort_by:'swidtag',sort_order:'asc',scopes:'#(scope)',search_params.equipment_id.filter_type:1,search_params.equipment_id.filteringkey:'#(data.Cluster.vcenter_name)'}
    When method get 
    Then status 200  


#-----------------------------Maintenance of a product of server--------------------------#
@get
Scenario: To get the maintance details of product of server
    Given path 'product','acqrights' 
    And params {page_num:1,page_size:50,sort_by:'PRODUCT_NAME',sort_order:'asc',scopes:'#(scope)',search_params.swidTag.filteringkey:'#(data.server.product_swidTag)',search_params.SKU.filteringkey:'null'}
   When method get
   Then status 200 

#--------------------Maintenance of a product of Softpartition--------------------------#
@get
Scenario: To get the maintance details of product of Softpartition

    Given path 'product','acqrights'
    And params {page_num:1,page_size:50,sort_by:'PRODUCT_NAME',sort_order:'asc',scopes:'#(scope)',search_params.swidTag.filteringkey:'#(data.Softpartition.product_swidTag)',search_params.SKU.filteringkey:'null'}
   When method get
   Then status 200 






  
   
    
     
    
    









   











  