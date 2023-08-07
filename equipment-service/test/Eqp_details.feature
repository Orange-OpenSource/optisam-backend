Feature: Equipment Service Test

Background:
# * def equipmentServiceUrl = "https://optisam-equipment-dev.apps.fr01.paas.tech.orange"
  * url equipmentServiceUrl+'/api/v1'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'
@get
Scenario: To get the items of server
    Given path 'equipment',data.Update_Eqp.parent_id ,'equipments'
    And params {page_num:1,page_size:50,sort_by:'server_id',sort_order:'asc',scopes:'#(scope)'}
    When method get 
    Then status 200
    And match response.totalRecords == 78

@sort
Scenario Outline: To check sorting  on server 
Given path 'equipment',data.Update_Eqp.parent_id ,'equipments'
And params { page_num:1, page_size:<page_size>,sort_by:'server_id',sort_order:<sortOrder>,scopes:'#(scope)'}
When method get
Then status 200 
And match response.totalRecords == 78

Examples:
|page_size| sortOrder|
|50|asc| 
|100|desc|



@sort
Scenario Outline: To check sorting  on server 
Given path 'equipment',data.Update_Eqp.parent_id ,'equipments'
And params { page_num:1, page_size:<page_size>,sort_by:'server_id',sort_order:<sortOrder>,scopes:'#(scope)'}
When method get
Then status 400

Examples:
|page_size| sortOrder|
|5|asc|
|"A"|desc|

@get
Scenario: To Get the Details of an Cluster 
    Given path 'equipment',data.equipmentID.cluster_id,'equipments'
    And  params { page_num:1,page_size:100,sort_by:'cluster_name',sort_order:'asc',scopes:'#(scope)'}
    When method get 
    Then status 200
    And match response.totalRecords == 14

@search
    #issue search name parsing 
Scenario Outline: To validate searching on Cluster 
    Given path 'equipment',data.equipmentID.cluster_id,'equipments'
    And  params { page_num:1,page_size:100,sort_by:'cluster_name',sort_order:'asc',scopes:'#(scope)',search_params:<cluster_name>}
    When method get 
    Then status 200
    And match response.totalRecords == 1
    Examples:
    |cluster_name|
    |cluster_name=cl01|
    |cluster_name=cl02|
    |cluster_name=cl03|
    

Scenario Outline: To validate searching on Cluster 
    Given path 'equipment',data.equipmentID.cluster_id,'equipments'
    And  params { page_num:1,page_size:100,sort_by:'cluster_name',sort_order:'asc',scopes:'#(scope)',search_params:<cluster_name>}
    When method get 
    Then status 200
    And match response.totalRecords == 0
    Examples:
    |cluster_name|
    |cluster_name=clxvd04|
    |cluster_name=cl0134r5345345|

#----------------------------------Vcenter-----------------------------------------------#

@get
Scenario: To get the detail of an particular Vcenter
    Given path 'equipment',data.equipmentID.vcenter_id,'equipments', data.Cluster.vcenter_name
    And  params { scopes:'#(scope)'}
    When method get 
    Then status 200
@get
Scenario:To get the parent of a particular Vcenter
    Given path 'equipment',data.equipmentID.vcenter_id, data.Cluster.vcenter_parentID,'parents'
    And  params { scopes:'#(scope)'}
    When method get 
    Then status 404
@get
Scenario:To get the Childrens  of a particular Vcenter
    Given path 'equipment',data.equipmentID.vcenter_id, data.Cluster.vcenter_parentID,'childs',data.equipmentID.cluster_id
    And  params { page_num:1,page_size:50,sort_by:'cluster_name',sort_order:'asc', scopes:'#(scope)'}
    When method get 
    Then status 200
    And match response.totalRecords == 2

@get
Scenario:To search the  Childrens  of a particular Vcenter
    Given path 'equipment',data.equipmentID.vcenter_id, data.Cluster.vcenter_parentID,'childs',data.Cluster.vcenter_childID
    And  params { page_num:1,page_size:50,sort_by:'cluster_name',sort_order:'asc', scopes:'#(scope)',search_params:'#(data.Cluster.vcenter_childname)'}
    When method get 
    Then status 404
    #And match response.totalRecords == 1
#---------------------------Editing or Updating the content of product------------------#
#Scenario: To edit the content of product - no of EqP User 
#    Given path 'equipment','allocatedmetric'
#   And request data.Eqp_product_Edit
#   When method put
#   Then status 200

#Scenario Outline: To edit the content of product - no of EqP User 
#    Given path 'equipment','allocatedmetric'
#    * set data.Eqp_product_Edit.equipment_user = <user>
#    * set data.Eqp_product_Edit.allocated_metrics = <metrics>
#   And request data.Eqp_product_Edit
#   When method put
#   Then status 200
#   Examples:
#   |user|metrics|
#   |6|""|
#   |0|"oracle.processor"|
#   |"15"|"oracle.nup"|
   

#Scenario Outline: To edit the content of product - no of EqP User -- 
#    Given path 'equipment','allocatedmetric'
#    * set data.Eqp_product_Edit.equipment_user = <user>
#    * set data.Eqp_product_Edit.allocated_metrics = <metrics>
#   And request data.Eqp_product_Edit
#   When method put
#   Then status 400

#   Examples:
#   |user|metrics|
#   |-6|""|
#   |" "|"oracle.processor"|
#   |"A"|"Random text"|
#   |"four"|"oracle.processor"|

   
   












  