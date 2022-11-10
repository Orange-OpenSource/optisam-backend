@aut @ignore @aut-setup

Feature: Pre-Requisite Setup for AUT(Automation) - Upload Data files

## Pre-requisite :  
# 1. equpiment type is created

  Background:
    * url dpsServiceUrl+'/api/v1'
    #* def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def scope = "AUT"

 Scenario: Upload Data files
    Given url importServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'import/data'
    * def file1_tmp = karate.readAsString('data/equipment_datacenter.csv')
    * def file2_tmp = karate.readAsString('data/equipment_vcenter.csv')
    * def file3_tmp = karate.readAsString('data/equipment_cluster.csv')
    * def file4_tmp = karate.readAsString('data/equipment_server.csv')
    * def file5_tmp = karate.readAsString('data/equipment_partition.csv')
    * def file6_tmp = karate.readAsString('data/products.csv')
    * def file7_tmp = karate.readAsString('data/applications.csv')
    * def file8_tmp = karate.readAsString('data/products_acquiredRights.csv')
    * def file9_tmp = karate.readAsString('data/applications_instances.csv')
    * def file10_tmp = karate.readAsString('data/applications_products.csv')
    * def file11_tmp = karate.readAsString('data/instances_equipments.csv')
    * def file12_tmp = karate.readAsString('data/instances_products.csv')
    * def file13_tmp = karate.readAsString('data/products_equipments.csv')
    * multipart file file = { value: '#(file1_tmp)', filename: 'equipment_datacenter.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file2_tmp)', filename: 'equipment_vcenter.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file3_tmp)', filename: 'equipment_cluster.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file4_tmp)', filename: 'equipment_server.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file5_tmp)', filename: 'equipment_partition.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file6_tmp)', filename: 'products.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file7_tmp)', filename: 'applications.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file8_tmp)', filename: 'products_acquiredRights.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file9_tmp)', filename: 'applications_instances.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file10_tmp)', filename: 'applications_products.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file11_tmp)', filename: 'instances_equipments.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file12_tmp)', filename: 'instances_products.csv', contentType: "text/csv" }
    * multipart file file = { value: '#(file13_tmp)', filename: 'products_equipments.csv', contentType: "text/csv" }
    * multipart field scope = scope
    When method post
    Then status 200 
    * call pause 20000
    * def today = todayDate('yyyy-MM-dd')
    Given url dpsServiceUrl+'/api/v1'
    Given path 'uploads/data'
    * header Authorization = 'Bearer '+access_token
    * params { page_num:1, page_size:10, sort_by:'upload_id', sort_order:'desc', scope:'#(scope)'}
    When method get
    Then status 200 
    * match response.uploads[0].uploaded_on contains today
    * match response.uploads[1].uploaded_on contains today
    * call pause 10000

## Verify acqRgihts data 
 Scenario: Verify Acquired Rights data
    Given url productServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    * path 'acqrights'
    * params {page_num:1, page_size:20, sort_by:'SWID_TAG', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    * match response.totalRecords == 21

  ## Verify Products data 
  Scenario: Verify Products Data
    Given url productServiceUrl+'/api/v1'
    * path 'products'
    * header Authorization = 'Bearer '+access_token
    * params {page_num:1, page_size:20, sort_by:'name', sort_order:'desc', scopes:'#(scope)'}
    When method get
    Then status 200
    * match response.totalRecords == 28


  ## Verify Equipments data
  Scenario: Verify Equipments Data
    * url equipmentServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments/types'
    * param scopes = scope
    When method get
    Then status 200
    * def partition_id =  karate.jsonPath(response.equipment_types, "$.[?(@.type=='partition')].ID")[0]
    * def server_id =  karate.jsonPath(response.equipment_types, "$.[?(@.type=='server')].ID")[0]
    * def cluster_id =  karate.jsonPath(response.equipment_types, "$.[?(@.type=='cluster')].ID")[0]
    * def vcenter_id =  karate.jsonPath(response.equipment_types, "$.[?(@.type=='vcenter')].ID")[0]
    * def datacenter_id =  karate.jsonPath(response.equipment_types, "$.[?(@.type=='datacenter')].ID")[0]
    * url equipmentServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments', partition_id,'equipments'
    * params { page_num:1, page_size:50, sort_by:'partition_code', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    # * match response.totalRecords == 5
    * assert response.equipments != 'W10='
    * url equipmentServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments', server_id,'equipments'
    * params { page_num:1, page_size:50, sort_by:'server_code', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    # * match response.totalRecords == 5
    ## Assert actual equipments data after decoding
    * assert response.equipments != 'W10='
    * url equipmentServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments', cluster_id,'equipments'
    * params { page_num:1, page_size:50, sort_by:'cluster_code', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    # * match response.totalRecords == 5
    ## Assert actual equipments data after decoding
    * assert response.equipments != 'W10='
        * url equipmentServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments', vcenter_id,'equipments'
    * params { page_num:1, page_size:50, sort_by:'vcenter_code', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    # * match response.totalRecords == 5
    * assert response.equipments != 'W10='
        * url equipmentServiceUrl+'/api/v1'
    * header Authorization = 'Bearer '+access_token
    Given path 'equipments', datacenter_id,'equipments'
    * params { page_num:1, page_size:50, sort_by:'datacenter_code', sort_order:'asc', scopes:'#(scope)'}
    When method get
    Then status 200
    # * match response.totalRecords == 5
    ## Assert actual equipments data after decoding
    * assert response.equipments != 'W10='