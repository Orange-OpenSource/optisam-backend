Feature: Group Compliance Service Test 

Background:
  * url productServiceUrl+'/api/v1/product'
  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}  
  * callonce read('common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = 'API'


Scenario: To get all the created Group 
    Given url "https://optisam-account-dev.apps.fr01.paas.tech.orange/api/v1/account"
    And path 'complience/groups'
    When method get 
    Then status 200 
    * match response.complience_groups[*].group_id contains data.complience_groups.group_id

Scenario: To view the Expanse Percentage,total cost,total expenditure for a group 
Given path 'dashboard/compliance/soft_exp'
And params { scope: 'OLN',scope: 'OCM',scope: 'CLR'}
When method get 
Then status 200 

Scenario: To get the Editor list in  the selected group 
    Given path "aggregations/editors"
    And params { scope: "OLN,OCM,CLR" }
    When method get 
    Then status 200 
    * match response.editor contains data.List_editor.editor

Scenario: To get the Expense By percent for the group selected.
    Given path 'dashboard/underusage'
    And params {"scopes":["OLN","OCM","CLR"]}
    And params { pageNum:1,pageSize:50,sortBy:'metrics',sortOrder:'asc' }
    When method get 
    Then status 200

Scenario: To View the group Counterfitting,group TotalCost,and group UnderUsaeCost
    Given path 'dashboard/groupcompliance/editor'
    And params {"scopes":["OLN","OCM","CLR"]}
    And params {"editor":'#(data.editor)'}
    When method get 
    Then status 200
    

Scenario: To View the group compliance of an selected editor
    Given path 'dashboard/groupcompliance/editor/product'
    And params {"scopes":["OLN","OCM","CLR"]}
    And params {"editor":'#(data.editor)'}
    When method get 
    Then status 200

Scenario: To View the group compliance of an selected editor
    Given path 'dashboard/groupcompliance/editor/product'
    And params {"scopes":["OLN","OCM","CLR"]}
    And params {"editor":'#(data.editor)'}
    And params {"editor":'#(data.editor_product)'}
    When method get 
    Then status 200

Scenario: To get the Underusage for a Group by editor 
   Given path 'dashboard/underusage'
    And params {"scopes":["OLN","OCM","CLR"]}
    And params {sortBy:'metrics', sortOrder: 'asc',editor: '#(data.editor_for_underusage)'}
    When method get 
    Then status 200 

Scenario: To view the Counterfitting cost,Total cost and Underusage cost by scopes
    Given path 'dashboard/groupcompliance/product'
    And params {"scopes":["OLN","OCM","CLR"]}
    And params {editor: '#(data.editor)',product_name: '#(data.editor_product)'}
    When method get 
    Then status 200 