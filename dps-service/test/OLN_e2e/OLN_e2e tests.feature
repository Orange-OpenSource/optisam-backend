@oln @ignore @e2e
Feature: E2E Data injection test for OLN Scope

## Pre-requisite :  
# 1. Nifi flow must be running
# 2. Equipment type (zone and server) should be present in OLN scope
# 3. dps cron timing should be less then 4 minutes

  Background:
    * url dpsServiceUrl+'/api/v1'
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def scope = "OLN"
    * def data = read('data.json')


  @oln
  Scenario: Verify end to end data upload for OLN scope using Globat Data files
    ## Delete Inventory
    Given path 'data',scope
    * header Authorization = 'Bearer '+access_token
    When method delete
    Then status 200
    * match $.success == true
    * call pause 2000
    ## Upload Global files
    Given url importServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'import/globaldata'
    * def file1_tmp = karate.readAsString('products_acquiredRights.csv')
    * def file2_tmp = karate.readAsString('servers.csv')
    * multipart file file = { value: '#(file1_tmp)', filename: 'products_acquiredRights.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file2_tmp)', filename: 'servers.csv', contentType: "text/csv" }
    * multipart field scope = scope
    When method post
    Then status 200 
    And match response contains 'file uploaded'
    * def today = todayDate('yyyy-MM-dd')
    Given url dpsServiceUrl+'/api/v1'
    Given path 'uploads/globaldata'
    * header Authorization = 'Bearer '+access_token
    * params { page_num:1, page_size:10, sort_by:'upload_id', sort_order:'desc' , scope : '#(scope)'}
    When method get
    Then status 200 
    * match response.uploads[0].uploaded_on contains today
    * match response.uploads[1].uploaded_on contains today
    ## verify data files processed via nifi are present in optisam
    * call pause 300000
    Given path 'uploads/data'
    * header Authorization = 'Bearer '+access_token
    * params { page_num:1, page_size:10, sort_by:'upload_id', sort_order:'desc', scope:'#(scope)'}
    When method get
    Then status 200 
    * match response.uploads[0].uploaded_on contains today
    * match response.uploads[1].uploaded_on contains today
    * call pause 5000


## Verify products data 
 Scenario: Verify Acquired Rights data
    Given url productServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    * path 'acqrights'
    * params {page_num:1, page_size:20, sort_by:'SWID_TAG', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    * match response.totalRecords == 5
    And match response.acquired_rights == data.acquired_rights

  ## Verify acqRgihts data 
  Scenario: Verify Products Data
    Given url productServiceUrl+'/api/v1'
    * path 'products'
    * header Authorization = 'Bearer '+access_token
    * params {page_num:1, page_size:20, sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    * match response.totalRecords == 4
    And match response.products == data.products


  ## Verify Equipments data
  Scenario: Verify Equipments Data
    * url equipmentServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments/types'
    * param scopes = scope
    When method get
    Then status 200
    * def zone_id =  karate.jsonPath(response.equipment_types, "$.[?(@.type=='zone')].ID")[0]
    * def server_id =  karate.jsonPath(response.equipment_types, "$.[?(@.type=='server')].ID")[0]
    * url equipmentServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments', zone_id,'equipments'
    * params { page_num:1, page_size:50, sort_by:'zone', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    # * match response.totalRecords == 5
    * assert response.equipments != 'W10='
    * url equipmentServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments', server_id,'equipments'
    * params { page_num:1, page_size:50, sort_by:'zone_serverhostname', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    # * match response.totalRecords == 5
    ## Assert actual equipments data after decoding
    * assert response.equipments != 'W10='
