@Editor
Feature: Catalog Editor Test

  Background:
    * url catalogServiceUrl  
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
   # * def scope = 'API'


    #@Schema
    #Scenario: Schema Validation for Editor
     # Given path 'catalog/editors'
     # And params {pageNum:1, sortBy:name, sortOrder:asc, pageSize:50}
     # * def schema = data.Schema_Editor
     # When method get
     # Then status 200
     # * match response.editors == '#[_ > 0] schema'

    @Create
    Scenario: Create Editor with valid name
        Given path 'api/v1/catalog/editor'
        And request data.create_editor
        When method post
        Then status 200
        * match response.name == data.create_editor.name
      
    @Create
      Scenario: Create Editor with existing name
        Given path 'api/v1/catalog/editor'
        And request data.create_editor
        When method post
        Then status 500 
        * match response.message == data.Duplicate_editor_message.message 

      Scenario: Create Editor with blank editor name
        Given path 'api/v1/catalog/editor'
        * set data.create_editor.name = data.Editor_Blank_input.name
        And request data.create_editor
        When method post
        Then status 500 
        * match response.message == data.Blank_Editor_Name_Message.message

      Scenario: Create Editor with special character and blank
        Given path 'api/v1/catalog/editor'
        * set data.create_editor.name = data.Editor_Blank_input_And_Special_Char.name
        And request data.create_editor
        When method post
        Then status 200
        * match response.name == data.Editor_Blank_input_And_Special_Char.name
 
        
      @UpdateEditor
      Scenario: Update created Editor details
        Given path 'catalog/editors'
        And params {pageNum:1, sortBy:productsCount, sortOrder:asc, pageSize:50,search_params.name.filteringkey:'#(data.Editor_Blank_input_And_Special_Char.name)'}
        When method get
        Then status 200
      * def editor_id = karate.jsonPath(response.editors,"$.[?(@.name=='"+data.Editor_Blank_input_And_Special_Char.name+"')]")[0].id
        * print  editor_id
        Given path 'api/v1/catalog/editor'
        * header Authorization = 'Bearer '+access_token
       * set data.Update_Editor_Information.id = editor_id
        And request data.Update_Editor_Information
        When method PUT
        Then status 200
        * match response.id == editor_id
        * match response.genearlInformation == data.Update_Editor_Information.genearlInformation

      Scenario: Delete created Editor
        Given path 'catalog/editors'
        And params {pageNum:1, sortBy:productsCount, sortOrder:asc, pageSize:50,search_params.name.filteringkey:'#(data.create_editor.name)'}
        When method get
        Then status 200
      * def editor_id = karate.jsonPath(response.editors,"$.[?(@.name=='"+data.create_editor.name+"')]")[0].id
        * print  editor_id
        Given path 'api/v1/catalog/editor' , editor_id
        * header Authorization = 'Bearer '+access_token
        When method delete
        Then status 200
        * match response.success == true
        * header Authorization = 'Bearer '+access_token
        Given path 'catalog/editors'
        And params {pageNum:1, sortBy:productsCount, sortOrder:asc, pageSize:50,search_params.name.filteringkey:'#(data.Update_Editor_Information.name)'}
        When method get
        Then status 200
      * def editor_id2 = karate.jsonPath(response.editors,"$.[?(@.name=='"+data.Update_Editor_Information.name+"')]")[0].id
        * print  editor_id
        Given path 'api/v1/catalog/editor' , editor_id2
        * header Authorization = 'Bearer '+access_token
        When method delete
        Then status 200
        * match response.success == true
        Given path 'catalog/editors'
        And params {pageNum:1, sortBy:productsCount, sortOrder:asc, pageSize:50,search_params.name.filteringkey:'#(data.Editor_Blank_input_And_Special_Char.name)'}
        When method get
        Then status 200
    

      Scenario: Get All editor name
        Given path '/catalog/editornames'
        When method get
        Then status 200
        * def editor_name = $response.editors[*].name

      @search
      Scenario Outline: To verify searching is working for Editor for <searchBy>
        Given path 'catalog/editors'
       And params {pageNum:1, sortBy:productsCount, sortOrder:desc ,pageSize:50}
        And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
        When method get
        Then status 200
        And match response.editors[*].<searchBy> contains '<searchValue>'
        Examples:
          | searchBy | | searchValue |
          | name | | Microsoft |
          | name | | Microsoft (Navision) |
        
        @search
        Scenario Outline: To verify searching for Editor with invalid editor <searchBy>
          Given path 'catalog/editors'
         And params {pageNum:1, sortBy:productsCount, sortOrder:desc ,pageSize:50}
          And params {search_params.<searchBy>.filteringkey: '<searchValue>'}
          When method get
          Then status 200
          
          And match response.totalrecords == 0 

          Examples:
            | searchBy | | searchValue |
            | name | | test123invalid |
        
@pagination
Scenario Outline: To verify pagination is working for editors

  Given path 'catalog/editors'
  And params {pageNum: 1,pageSize: <pageSize>,sortBy: name,sortOrder: asc}
  When method get
  Then status 200
  And response.totalRecords > 0
  And match $.editors == '#[_ <= <pageSize>]'
  
  Examples:
  | pageSize |
  | 50 |
  | 100 |
  | 200 |

@pagination
  Scenario Outline: To verify Invalid pagination input for editors
  
      Given path 'catalog/editors'
      And params {pageNum: 1,pageSize: <pageSize>,sortBy: name,sortOrder: asc}
      When method get
      Then status <code>
      And response.totalrecords == 0
      
      Examples:
      | pageSize |code|
      | "ABC" |405|
      | 1002 |405|
      | "100" | 200|          