@Product
Feature: Catalog Product Test

  Background:
    * url catalogServiceUrl
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')


    #Scenario: Schema Validation for Product
    #Given path 'catalog/products'
    #And params {page_num:1, page_size:50, sort_by:createdOn, sort_order:asc}
    #* def schema = data.Product_Schema
    #When method get
    #Then status 200
    #* match response.product == '#[_ > 0] schema'

    Scenario: Create product with valid details
       #Create Editor
        Given path 'api/v1/catalog/editor'
        And request data.create_editor
        When method post
        Then status 200
        * match response.name == data.create_editor.name
       # Get the editor ID
        Given path 'catalog/editors'
        * header Authorization = 'Bearer '+access_token
        And params {pageNum:1, sort_by:name, sort_order:asc, pageSize:50,search_params.name.filteringkey: '#(data.create_editor.name)' }
        When method get
        Then status 200
      * def editor_id = karate.jsonPath(response.editors,"$.[?(@.name=='"+data.create_editor.name+"')]")[0].id
        * print  editor_id
        # Create Product
        Given path 'api/v1/catalog/product'
        * header Authorization = 'Bearer '+access_token
       * set data.Create_Product.editorID = editor_id
       * set data.Create_Product.editorName = data.create_editor.name
       And request data.Create_Product
       When method post
       Then status 200
       * match response.editorID == editor_id
       * match response.name == data.Create_Product.name

      Scenario: To verify duplicate product name is not allowed for same editor
        Given path 'catalog/editors'
        And params {pageNum:1, sortBy:createdOn, sortOrder:desc, pageSize:50}
        When method get
        Then status 200
      * def editor_id1 = karate.jsonPath(response.editors,"$.[?(@.name=='"+data.create_editor.name+"')]")[0].id
       * print editor_id1
        Given path 'api/v1/catalog/product'
        * header Authorization = 'Bearer '+access_token
       * set data.Create_Product.editorID = editor_id1
       And request data.Create_Product
       When method post
       Then status 500

      Scenario: Create product with empty product name
        Given path 'catalog/editors'
        And params {pageNum:1, sortBy:createdOn, sortOrder:desc, pageSize:50}
        When method get
        Then status 200
      * def editor_id1 = karate.jsonPath(response.editors,"$.[?(@.name=='"+data.create_editor.name+"')]")[0].id
        Given path 'api/v1/catalog/product'
        * header Authorization = 'Bearer '+access_token
       * set data.Create_Product.editorID = editor_id1
       * set data.Create_Product.name = data.Empty_Product_Name.name
       And request data.Create_Product
       When method post
       Then status 500

      Scenario: Create product with containg special character and space
        Given path 'catalog/editors'
        And params {pageNum:1, sortBy:createdOn, sortOrder:desc, pageSize:50}
        When method get
        Then status 200
      * def editor_id1 = karate.jsonPath(response.editors,"$.[?(@.name=='"+data.create_editor.name+"')]")[0].id
        Given path 'api/v1/catalog/product'
        * header Authorization = 'Bearer '+access_token
       * set data.Create_Product.editorID = editor_id1
       * set data.Create_Product.name = data.Product_Name_With_Special_character.name
       And request data.Create_Product
       When method post
       Then status 200
        * match response.name == data.Product_Name_With_Special_character.name


      Scenario: Create product with multiple vandor, audits, etc
        Given path 'catalog/editors'
        And params {pageNum:1, sortBy:createdOn, sortOrder:desc, pageSize:50}
        When method get
        Then status 200
      * def editor_id1 = karate.jsonPath(response.editors,"$.[?(@.name=='"+data.create_editor.name+"')]")[0].id
        Given path 'api/v1/catalog/product'
        * header Authorization = 'Bearer '+access_token
       * set data.Create_Product_With_Multiple_Input.editorID = editor_id1
       And request data.Create_Product_With_Multiple_Input
       When method post
       Then status 200
        * match response.name == data.Create_Product_With_Multiple_Input.name
     
      Scenario: Update product using PUT call
        #getting editor id
        Given path 'catalog/editors'
        And params {pageNum:1, sortBy:createdOn, sortOrder:desc, pageSize:50,search_params.name.filteringkey: '#(data.create_editor.name)' }
        When method get
        Then status 200
        * def editor_id1 = karate.jsonPath(response.editors,"$.[?(@.name=='"+data.create_editor.name+"')]")[0].id
       * print 'editor_id1..' + editor_id1
         # getting product ID
         * def sleep = function(millis){ java.lang.Thread.sleep(millis) }
         * sleep(1000)
        Given path 'catalog/products'
        * header Authorization = 'Bearer '+access_token
        And params {page_num:1, page_size:200, sort_by:createdOn, sort_order:desc,search_params.name.filteringkey:'#(data.Create_Product.name)'}
        When method get
        Then status 200
        * match response.total_records == '#number? _ >= 0' 
        * def product_id = karate.jsonPath(response.product,"$.[?(@.editorName=='"+data.create_editor.name+"')]")[0].id
        * print 'product_id....' + product_id
        #Updating product using PUT
        Given path 'api/v1/catalog/product'
        * header Authorization = 'Bearer '+access_token
       * set data.Update_Product_Details.id = product_id
       * set data.Update_Product_Details.editorID = editor_id1
       And request data.Update_Product_Details
       When method PUT
       Then status 200
       * match response.id == product_id
       * match response.name == data.Update_Product_Details.name
       * match response.genearlInformation == data.Update_Product_Details.genearlInformation
        
        @get
        Scenario: Get Single product details 
            #Getting product ID
            Given path 'catalog/products'
            * header Authorization = 'Bearer '+access_token
            And params {page_num:1, page_size:50, sort_by:createdOn, sort_order:desc,search_params.name.filteringkey:'#(data.Create_Product.name)'}
            When method get
            Then status 200
            

            Scenario: Delete the product
                #Getting product ID
            Given path 'catalog/products'
            * header Authorization = 'Bearer '+access_token
            And params {page_num:1, page_size:50, sort_by:name, sort_order:desc,search_params.name.filteringkey:'#(data.delete_product.name)'}
            When method get
            Then status 200
            * match response.total_records == '#number? _ >= 0'
            * def product_id = karate.jsonPath(response.product,"$.[?(@.name=='"+data.delete_product.name+"')]")[0].id
            * print 'product_id....' + product_id
             #Getting single product 
            Given path 'api/v1/catalog/product', product_id
            * header Authorization = 'Bearer '+access_token 
            When method Delete
            Then status 200
            * match response.success == true

        Scenario: Delete created Editor
            Given path 'catalog/editors'
            And params {pageNum:1,  pageSize:50, sortBy:createdOn, sortOrder:desc}
            When method get
            Then status 200
          * def editor_id = karate.jsonPath(response.editors,"$.[?(@.name=='"+data.create_editor.name+"')]")[0].id
            * print  editor_id
            Given path 'api/v1/catalog/editor' , editor_id
            * header Authorization = 'Bearer '+access_token
            When method delete
            Then status 200
            * match response.success == true

 
Scenario Outline: To verify searching is working for products for <searchBy>
  Given path 'catalog/products'
  And params {page_num:1, page_size:50, sort_by:name, sort_order:asc}
  And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
  When method get
  Then status 200
  And match response.product[*].<searchBy> contains '<searchValue>'
  
  Examples:
  | searchBy | | searchValue |
  | name | | Adobe Reader |
  | editorName | | Adobe | 
  

  
Scenario Outline: To verify searching is working for locationType <searchBy> locationType
  Given path 'catalog/products'
  And params {page_num:1, page_size:50, sort_by:name, sort_order:asc}
  And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
  When method get
  Then status 200
  
  And match response.product[*].<searchBy> contains '<searchValue>'
  
  
  Examples:
  | searchBy | | searchValue |
  | locationType | | SAAS |
  

Scenario Outline: To verify searching is working for  <searchBy> 
  Given path 'catalog/products'
  And params {page_num:1, page_size:50, sort_by:name, sort_order:asc}
  And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
  When method get
  Then status 200
  And match response.product[*].licensing contains [<searchValue>]
  
  Examples:
  | searchBy | | searchValue |
  | licensing | |OPENSOURCE |
  
Scenario Outline: To verify searching is working for  <searchBy> 
  Given path 'catalog/products'
  And params {page_num:1, page_size:50, sort_by:name, sort_order:asc}
  And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
  When method get
  Then status 200
  And match response.product[*].licensing contains [<searchValue>]
  Examples:
  | searchBy | | searchValue |
  | licensing | | CLOSEDSOURCE |
  
Scenario Outline: To verify searching for invalid <searchBy> input
  Given path 'catalog/products'
  And params {page_num:1, page_size:50, sort_by:name, sort_order:asc}
  And params {search_params.<searchBy>.filteringkey: <searchValue>}
  When method get
  Then status 200
 #And match response.total_records == 0
  
  Examples:
  | searchBy | | searchValue |
  | locationType | | sdcvdvvxcvv |
  | licensing | | fgdhdfghfzdvxvgj |


Scenario Outline: To verify searching product with more than one combination of product and editor <searchBy1> and <searchBy2> 
Given path 'catalog/products'
And params {page_num:1, page_size:50, sort_by:name, sort_order:asc}
  And params {search_params.<searchBy1>.filteringkey: '<searchValue1>'}
  And params {search_params.<searchBy2>.filteringkey: '<searchValue2>'}
  When method get
  Then status 200
  And response.totalRecords > 0

  And match response.product[*].<searchBy1> contains '<searchValue1>'
  And match response.product[*].<searchBy2> contains '<searchValue2>'

  Examples:
  | searchBy1 | searchValue1 | searchBy2 | searchValue2 |
 | name | 3CServer | editorName | Hewlett-Packard (3Com) |

Scenario Outline: To verify searching product with more than one invalid combination of product and editor <searchBy1> and <searchBy2> 
  Given path 'catalog/products'
  And params {page_num:1, page_size:50, sort_by:name, sort_order:asc}
      And params {search_params.<searchBy1>.filteringkey: '<searchValue1>'}
      And params {search_params.<searchBy2>.filteringkey: '<searchValue2>'}
      When method get
      Then status 200
      And response.totalRecords == 0
      Examples:
      | searchBy1 | searchValue1 | searchBy2 | searchValue2 |
      | name | 3CServer | data.Search_record.Search_Valid_Product_Name | data.Search_record.Search_Invalid_EditorName | 

  Scenario Outline: To verify searching product with more than one combination of location and licensing <searchBy1> and <searchBy2> 
      Given path 'catalog/products'
      And params {page_num:1, page_size:50, sort_by:name, sort_order:asc}
          And params {search_params.<searchByLicensing>.filteringkey: '<searchLicensingValue1>'}
          And params {search_params.<searchBylocationType>.filteringkey: '<searchlocationTypeValue2>'}
          When method get
          Then status 200
          And response.totalRecords > 0
          And match response.product[*].<searchByLicensing> contains [<searchLicensingValue1>]
          And match response.product[*].<searchBylocationType> contains '<searchlocationTypeValue2>'
      
          Examples:
          | searchByLicensing | searchLicensingValue1 | searchBylocationType | searchlocationTypeValue2 |
          |      licensing    |       CLOSEDSOURCE    |      locationType    |           SAAS           |           

        @pagination
        Scenario Outline: To verify pagination is working for editors

          Given path 'catalog/products'
          And params {page_num:1, page_size:<pageSize>, sort_by:name, sort_order:asc}
          When method get
          Then status 200
          And response.total_records > 0
          And match $.product == '#[_ <= <pageSize>]'
          
          Examples:
          | pageSize |
          | 50 |
          | 100 |
          | 200 |        