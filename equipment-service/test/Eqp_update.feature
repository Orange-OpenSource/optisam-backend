Feature: Equipment Service Test

Background:
# * def equipmentServiceUrl = "https://optisam-equipment-int.apps.fr01.paas.tech.orange"
  * url equipmentServiceUrl+'/api/v1/equipment'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'


@get
Scenario:Get Details of All Equipment - Softpartition
 Given path data.equipmentID.softpartition_id, 'equipments'
And params { page_num:1, page_size:50, sort_by:'softpartition_name', sort_order:'desc', scopes:'#(scope)'}
When method get 
Then status 200
And match response.totalRecords == 25
@get
   
@get
Scenario:Get Details of a particular  Equipment - Softpartition
    Given path data.equipmentID.softpartition_id, 'equipments', data.Softpartition.ID
   And params {  scopes:'#(scope)'}
   When method get 
   Then status 200
   And response.vcpu ==4
@get
Scenario:Get Details of parent of Equipment - Softpartition
    Given path data.equipmentID.softpartition_id, data.Softpartition.softpartition_parent_id,'parents'
   And params {  scopes:'#(scope)'}
   When method get 
   Then status 404
   And match response.message == "Equipment Parent doesn't exists"




@search
Scenario: To check the Searching on Equipment -Softpartition in Advance Search with single input
    Given path data.equipmentID.softpartition_id ,'equipments'
    And params { page_num:1,page_size:50,sort_by:'softpartition_id',sort_order:asc,scopes:'#(scope)' }
   And params { search_params:'#(data.Softpartition.ID)'}
    When method get 
    Then status 400
@search
Scenario: To check the Searching on Equipment -Softpartition in Advance Search with double  input
    Given path data.equipmentID.softpartition_id ,'equipments'
    And params { page_num:1,page_size:50,sort_by:'softpartition_id',sort_order:asc,scopes:'#(scope)',search_params:softpartition_name='#(data.Softpartition.Softpartition_Name)',softpartition_id:'#(data.Softpartition.ID)'}
    When method get 
    Then status 200
@search
Scenario: To check the Searching on Equipment -Softpartition with Invalid Input
    Given path data.equipmentID.softpartition_id ,'equipments'
    And params { page_num:1,page_size:50,sort_by:'softpartition_id',sort_order:asc,scopes:'#(scope)',search_params:'Worng Input'}
    When method get 
    Then status 400

#------------------------------------------Cluster-------------------------------------------------
@get
Scenario:Get Details of All Equipment - Cluster
    Given path data.equipmentID.cluster_id, 'equipments'
   And params { page_num:1, page_size:50, sort_by:'cluster_name', sort_order:'desc', scopes:'#(scope)'}
   When method get 
   Then status 200 
   
@get
Scenario:Get Details of a Individual Equipment - Cluster
    Given path data.equipmentID.cluster_id, 'equipments', data.Cluster.cluster_name
   And params { scopes:'#(scope)'}
   When method get 
   Then status 200

 









   





