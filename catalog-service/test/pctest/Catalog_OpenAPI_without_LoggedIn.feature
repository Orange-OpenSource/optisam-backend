@OpenAPIURL
Feature: Catalog Editor Test

Background:
  * url catalogServiceUrl
  #* url ProductcatalogServiceUrl
  * def data = read('data.json')


Scenario: Schema validation for editor
    Given path 'catalog/editors' 
    And params {pageNum:1, pageSize:50, sortBy:name, sortOrder:desc}
    * def schema = data.Schema_OpenAPI_Editor
    When method get
    Then status 200
    * response.totalrecords == '#number? _ >= 0'
    #* match response.editors == '#[_ > 0] schema'
    #* match response.editors == '#[_ <= 50] schema'

@get
Scenario: To verify Open URL Get for list of all editors
    Given path 'catalog/editors'
    And params {pageNum:1, sortBy:createdOn, sortOrder:asc ,pageSize:50}
    When method get
    Then status 200


Scenario: To Get the all filters for Editors
    Given path 'catalog/editorfilters'
    When method get 
    Then status 200 
    * match response.groupContract.total_count == data.Group_Contract.total_count
    * match response.year.total_count == data.year.total_count
    * match response.countryCode.total_count == data.countryCode.total_count
    * match response.entities.total_count == data.entities.total_count
    

Scenario Outline: To Verify the result after applying the filters(Group Contract)
    Given path 'catalog/editors'
    And params {pageNum: 1, pageSize: 50, sortBy: 'name', sortOrder: 'asc', search_params.group_contract.filteringkey: <Selection>}
    When method get
    Then status 200
    * response.totalrecords == '#number? _ >= 0'
    * match response.editors[*].groupContract contains [ <Selection> ]
    
    Examples:
      | Selection    |
      | true         |
      | false        |
      


Scenario Outline: To Verify the result after applying the filters(Entities)
    Given path 'catalog/editors'
    And params {pageNum: 1,pageSize: 50,sortBy: name,sortOrder: asc,search_params.entities.filteringkey: <Entity>}
    When method get 
    Then status 200 
    * response.totalrecords == '#number? _ >= 0'
    * match response.editors[*].scopes[*] contains [ <Entity> ]
    Examples:
    |Entity|
    |AAA|
    |Akanksha|
    |Aakash|
    |CHA|


Scenario Outline: To Verify the result after applying the filters(Audit Year)
    Given path 'catalog/editors'
    And params {pageNum: 1,pageSize: 50,sortBy: name,sortOrder: asc,search_params.audityears.filteringkey: <AuditYear>}
    When method get 
    Then status 200 
    * response.totalrecords == '#number? _ >= 0'
    * match response.editors[*].audits[*].year contains [ <AuditYear> ]
    Examples:
    |AuditYear|
    |2026|
    |2023|
    |2022|
    |2021|

Scenario Outline: To Verify the result after applying the filters(Country)
    Given path 'catalog/editors'
    And params {pageNum: 1,pageSize: 50,sortBy: name,sortOrder: asc,search_params.countryCodes.filteringkey:<Country>}
    When method get 
    Then status 200 
    * response.totalrecords == '#number? _ >= 0'
    * match response.editors[*].country_code contains [ <Country> ]
    Examples:
    |Country|
    |al|
    |in|
    |aq|
    |az|

Scenario Outline: To Verify the result after applying multiple filters
    Given path 'catalog/editors'
    And params { pageNum: 1,pageSize: 50,sortBy: name,sortOrder: asc,search_params.audityears.filteringkey: <AuditYear>,search_params.countryCodes.filteringkey: <Country_Codes>,search_params.entities.filteringkey: <Entities>,search_params.group_contract.filteringkey: <Group_Contract> }
     When method get 
    Then status 200 
    * response.totalrecords == '#number? _ >= 0'
    * match response.editors[*].groupContract contains [ <Group_Contract> ]
    * match response.editors[*].scopes[*] contains [ <Entities> ]
    * match response.editors[*].audits[*].year contains [ <AuditYear> ]
    * match response.editors[*].country_code contains [ <Country_Codes> ]
  Examples:
  |AuditYear|Country_Codes|Entities|Group_Contract|
  |2022|bv|Aakash|true|
  
Scenario Outline: To verify the Function of Searching  and Pagination in the Editor
    Given path 'catalog/editors'
    And params {pageNum: 1,pageSize:<page_size>,sortBy: name,sortOrder: asc,search_params.name.filteringkey:<name>}
    When method get 
    Then status <code>
    Examples:
    |name|page_size|code|
    |'Deepak'|50|200|
    |'Adobe'|100|200|
    |'Oracle'|200|200|
    |'ruitpioerutepirtu'|"a"|405|
    |'tyiyityiooyutryurtpyouri'|"SFSDF"|405|
    |'Deepak'|12|200|

Scenario: To verify the clicking on the Editor Tile 
    Given path 'catalog/products'
    And params {page_num: 1,page_size: 50,sort_by: name,sort_order: asc,search_params.editorId.filteringkey:'#(data.Editor.ID)'}
    When method get 
    Then status 200 
    * match response.product[*].editorID contains  data.Editor.ID

Scenario:To verfiy the action after clicking on the product name in tha editor view
    Given path 'catalog/products'
    And params {page_num: 1,page_size: 50,sort_by: name,sort_order: asc}
    When method get 
    Then status 200 
    * match response.product[*].editorID contains  data.Editor.ID


Scenario: To verify the searching in the product view of Editor
    Given path 'catalog/products'
    And params {page_num: 1,page_size: 50,sort_by: name,sort_order: asc,search_params.editorId.filteringkey:'#(data.Editor.ID)',search_params.name.filteringkey:'#(data.Editor.product_name)'}
    When method get 
    Then status 200 
    * match response.product[*].name contains data.Editor.product_name

Scenario: To get all the filters for Products
    Given path 'catalog/productfilters'
    When method get 
    Then status 200 
    * response.totalrecords == '#number? _ >= 0'
    * match response.deploymentType.total_count == data.deploymentType.total_count
    * match response.licensing.total_count == data.licensing.total_count
    * match response.recommendation.total_count == data.recommendation.total_count
    * match response.entities.total_count == data.Entity.total_count
    * match response.vendors.total_count == data.vendors.total_count

Scenario Outline: To verify the result after applying ther filters(Deployment Type)
    Given path 'catalog/products'
    And params {page_num: 1,page_size: 50,sort_by: name,sort_order: asc,search_params.deploymentType.filteringkey:<DeploymentType>}
    When method get 
    Then status 200 
    * response.totalrecords == '#number? _ >= 0'
    * match response.product[*].locationType contains [<DeploymentType>]
    Examples:
    |DeploymentType|
    |On Premise|
    |SAAS|
    |NONE|

Scenario Outline: To verify the result after applying ther filters(Entity)
    Given path 'catalog/products'
    And params {page_num: 1,page_size: 50,sort_by: name,sort_order: asc,search_params.entities.filteringkey:<Entity>}
    When method get 
    Then status 200 
    * response.totalrecords == '#number? _ >= 0'
    * match response.product[*].scopes[*] contains [<Entity>]
    Examples:
    |Entity|
    |AAA|
    |Akanksha|
    |Aakash|
    |CHA|

Scenario Outline: To verify the result after applying ther filters(Recommendation)
    Given path 'catalog/products'
    And params {page_num: 1,page_size: 50,sort_by: name,sort_order: asc,search_params.recommendation.filteringkey:<Recommendation>}
    When method get 
    Then status 200 
    * response.totalrecords == '#number? _ >= 0'
    * match response.product[*].recommendation contains [<Recommendation>]
    Examples:
    |Recommendation|
    |AUTHORIZED|
    |BLACKLISTED|
    |NONE|
    |RECOMMENDED|
    #|'AUTHORIZED,BLACKLISTED,NONE,RECOMMENDED'|


Scenario Outline: To verify the result after applying ther filters(Licensing)
    Given path 'catalog/products'
    And params {page_num: 1,page_size: 50,sort_by: name,sort_order: asc,search_params.licensing.filteringkey:<Licensing>}
    When method get 
    Then status 200 
    * response.totalrecords == '#number? _ >= 0'
    * match response.product[*].licensing contains [<Licensing>]
    Examples:
    |Licensing|
    |CLOSEDSOURCE|
    |NONE|
    |OPENSOURCE|
    #|'CLOSEDSOURCE,NONE,OPENSOURCE'|

Scenario Outline: To verify the result after applying ther filters(Vendors)
    Given path 'catalog/products'
    And params {page_num: 1,page_size: 50,sort_by: name,sort_order: asc,search_params.vendor.filteringkey:<vendors>}
    When method get 
    Then status 200 
    * response.totalrecords == '#number? _ >= 0'
    * match response.product[*].supportVendors[*] contains [<vendors>]
    Examples:
    |vendors|
    |'DEEPAK JASWAL'|
    |MSD|
    |s1|

Scenario: To vefiy the clicking on Product Tile
# No API get called on clicking this 

Scenario Outline: To verfiy searching in the Product Tab 
    Given path 'catalog/products'
    And params {page_num: 1,page_size: 50,sort_by: name,sort_order: asc,search_params.name.filteringkey:<name>}
    When method get 
    Then status 200 
    Examples:
    |name|
    |Deepak|
    |product|
    |oracle|

Scenario Outline: To verfiy Pagination  in the Product Tab 
    Given path 'catalog/products'
    And params {page_num: 1,page_size:<pageSize>,sort_by: name,sort_order: asc,search_params.name.filteringkey:name}
    When method get 
    Then status 200 
    * response.totalrecords == <pageSize>
    Examples:
    |pageSize|
    |50|
    |100|
    |200|


Scenario Outline: To verfiy Pagination  in the Product Tab --Invalid Input
    Given path 'catalog/products'
    And params {page_num: 1,page_size:<pageSize>,sort_by: name,sort_order: asc,search_params.name.filteringkey:name}
    When method get 
    Then status <code>
    Examples:
    |pageSize|code|
    |54|200|
    |"aa"|405|
    |"a"|405|








    









    
#    
#@get
#Scenario: To verify Open URL get single editor 
#    Given path 'catalog/editors'
#    And params {pageNum:1, sortBy:productsCount, sortOrder:desc ,pageSize:50}
#    When method get
#    Then status 200
#    * response.totalrecords > 0
#    * def result = karate.jsonPath(response, "$.editors[?(@.name=='"+data.Open_API_Get_Editor.name+"')]")[0].id
#   Given path 'catalog/editor'
#   #And params {pageNum:1, sortBy:createdOn, sortOrder:asc ,pageSize:50}
#   And params {id:'#(result)'}
#   When method get
#   Then status 200
#  * match response.scopes contains data.Open_API_Get_Editor.scope
# 
#@get
#Scenario: To verify Open URl GET for list of products
#    Given path 'catalog/products'
#    And params {page_num:1, page_size:50, sort_by:createdOn, sort_order:asc}
#    When method get
#    Then status 200
#
#Scenario Outline: To verify searching is working for products for <searchBy>
#    Given path 'catalog/products'
#    And params {page_num:1, page_size:50, sort_by:createdOn, sort_order:asc}
#    And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
#    When method get
#    Then status 200
#    And match response.product[*].<searchBy> contains '<searchValue>'
#    
#    Examples:
#    | searchBy | | searchValue |
#    | name | | Adobe Reader |
#    | editorName | | Adobe | 
#    
#Scenario Outline: To verify searching is working for locationType <searchBy> locationType
#    Given path 'catalog/products'
#    And params {page_num:1, page_size:50, sort_by:name, sort_order:asc}
#    And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
#    When method get
#    Then status 200
#    #* def matchValue = karate.lowerCase(response.product[*].<searchBy>) 
#    And match response.product[*].<searchBy> contains '<searchValue>'
#    #And match matchValue contains '<searchValue>'
#    
#    Examples:
#    | searchBy | | searchValue |
#    | locationType | | SAAS |
#    #| locationType | | saas |
#    #| locationType | | Open Premise  |
#    | locationType | | Both  |
#
#Scenario Outline: To verify searching is working for licensing <searchBy> licensing
#    Given path 'catalog/products'
#    And params {page_num:1, page_size:50, sort_by:createdOn, sort_order:asc}
#    And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
#    When method get
#    Then status 200
#    And match response.product[*].openSource.isOpenSource contains true
#    And match response.product[*].closeSource.isCloseSource contains false
#
#    Examples:
#    | searchBy | | searchValue |
#    | licensing | | Open Source |
#    
#Scenario Outline: To verify searching is working for licensing <searchBy> licensing
#    Given path 'catalog/products'
#    And params {page_num:1, page_size:50, sort_by:createdOn, sort_order:asc}
#    And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
#    When method get
#    Then status 200
#    And match response.product[*].closeSource.isCloseSource contains true
#    And match response.product[*].openSource.isOpenSource contains false
#    
#    Examples:
#    | searchBy | | searchValue |
#    | licensing | | Closed Source |
#    
#Scenario Outline: To verify searching for invalid locationType <searchBy> invalid input of locationType
#    Given path 'catalog/products'
#    And params {page_num:1, page_size:50, sort_by:createdOn, sort_order:asc}
#    And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
#    When method get
#    Then status 200
#   And match response.total_records == 0
#    
#    Examples:
#    | searchBy | | searchValue |
#    | locationType | | fgdhdfghfgj |
#    | licensing | | fgdhdfghfgj |
#
#
#Scenario Outline: To verify searching product with more than one combination of product and editor <searchBy1> and <searchBy2> 
#Given path 'catalog/products'
#And params {page_num:1, page_size:50, sort_by:createdOn, sort_order:asc}
#    And params {search_params.<searchBy1>.filteringkey: '<searchValue1>'}
#    And params {search_params.<searchBy2>.filteringkey: '<searchValue2>'}
#    When method get
#    Then status 200
#    And response.totalRecords > 0
#
#    And match response.product[*].<searchBy1> contains '<searchValue1>'
#    And match response.product[*].<searchBy2> contains '<searchValue2>'
#
#    Examples:
#    | searchBy1 | searchValue1 | searchBy2 | searchValue2 |
#   | name | 3CServer | editorName | Hewlett-Packard (3Com) |
#
#Scenario Outline: To verify searching product with more than one invalid combination of product and editor <searchBy1> and <searchBy2> 
#    Given path 'catalog/products'
#    And params {page_num:1, page_size:50, sort_by:createdOn, sort_order:asc}
#        And params {search_params.<searchBy1>.filteringkey: '<searchValue1>'}
#        And params {search_params.<searchBy2>.filteringkey: '<searchValue2>'}
#        When method get
#        Then status 200
#        And response.totalRecords == 0
#    
#        Examples:
#        | searchBy1 | searchValue1 | searchBy2 | searchValue2 |
#        | name | 3CServer | data.Search_record.Search_Valid_Product_Name | data.Search_record.Search_Invalid_EditorName | 
#
#    Scenario Outline: To verify searching product with more than one combination of location and licensing <searchBy1> and <searchBy2> 
#        Given path 'catalog/products'
#        And params {page_num:1, page_size:50, sort_by:name, sort_order:asc}
#            And params {search_params.<searchByLicensing>.filteringkey: '<searchLicensingValue1>'}
#            And params {search_params.<searchBylocationType>.filteringkey: '<searchlocationTypeValue2>'}
#            When method get
#            Then status 200
#            And response.totalRecords > 0
#
#            And match response.product[*].closeSource.isCloseSource contains true
#            And match response.product[*].<searchBylocationType> contains '<searchlocationTypeValue2>'
#        
#            Examples:
#            | searchByLicensing | searchLicensingValue1 | searchBylocationType | searchlocationTypeValue2 |
#           | licensing | Closed Source | locationType | SAAS |
#         
#
# #Negative test cases
#@get
#Scenario:  To verify GET URL for List of products without Params
#    Given path 'catalog/products'
#    #And params {page_num:1, page_size:50, sort_by:createdOn, sort_order:asc}
#    When method get
#    Then status 200 
#    * match response.total_records == 0
#
#@get
#Scenario: To verify Open URL GET single product
#    Given path 'catalog/products'
#    And params {page_num:1, page_size:500, sort_by:name, sort_order:asc}
#    #And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
#    When method get
#    Then status 200
#    * match response.total_records == '#number? _ >= 0'
#    * def result = karate.jsonPath(response, "$.product[?(@.name=='"+data.Open_API_Get_Product.name+"')]")[0].id
#  
#    Given path 'catalog/product'
#    And params {id:'#(result)'}
#    When method get
#    Then status 200
#    * print response
#    * match response.editorName == data.Open_API_Get_Product.editorName
#
#
#@get
#Scenario: To verify single product by passing invalid id
#    Given path 'catalog/product'
#    And params {id:'#(data.Open_API_Get_Product.invalid_id)'}
#    When method get
#    Then status 400
#
#
#Scenario Outline: To verify pagination is working for editors
#
#Given path 'catalog/editors'
#And params {page_num:1, page_size:<pageSize>, sort_by:productsCount, sort_order:desc}
#When method get
#Then status 200
#And response.totalRecords > 0
#And match $.editors == '#[_ <= <pageSize>]'
#
#Examples:
#| pageSize |
#| 50 |
#| 100 |
#| 200 |
#
#Scenario Outline: To verify Invalid pagination input for editors
#
#    Given path 'catalog/editors'
#    And params {page_num:1, page_size:<pageSize>, sort_by:productsCount, sort_order:desc}
#    When method get
#    Then status 200
#    And response.totalrecords == 0
#    
#    Examples:
#    | pageSize |
#    | "ABC" |
#    | 1002 |
#    | 10 |
#
##@sort   
# #   Scenario Outline: To verify Sorting is working on list of Editor by <sortBy>
#   
#  #  Given path 'catalog/editors'
#   # And params {page_num:1, page_size:50, sort_by:<sortBy>, sort_order:<sortOrder>}
#    #When method get
#    #Then status 200
#    #And response.totalrecords > 0
#    #* def actual = $response.editors[*].<sortBy>
#    #* def sorted = sort(actual,'<sortOrder>')
#    #* match sorted == actual
#
#    #Examples:
#    #| sortBy | sortOrder |  
#    #| productsCount | desc |
#   
#