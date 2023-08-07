@FileUpload
Feature:- To verify Upload functionality 

Background: 
    * url 'https://optisam-catalog-int.apps.fr01.paas.tech.orange/api/v1/'
    #* url catalogServiceUrlImport+ '/api/v1/'
    * def credentials = {username:'admin@test.com', password: 'Welcome@123'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')


    Scenario: Schema validation
        Given path 'catalog/bulkfileuploadlogs'
        When method get
        * def schema = data.Upload_Schema
        Then status 200
        * match response.uploadCatalogDataLogs == '#[_ > 0] schema'

    Scenario: To verify Upload API should work
        Given path 'catalog/bulkfileuploadlogs'
        When method get
        Then status 200


    #Scenario: To verify upload functionality for product catalog
        # Given path 'import/uploadcatalogdata'
        # * def file1_tmp = karate.readAsString('Catalog-TestData-import.xlsx')
        # * multipart file file = { read: 'this:data/Catalog-TestData-import.xlsx', filename: 'Catalog-TestData-import.xlsx', contentType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }
        # * multipart file file = { read: 'C:/API-Automation/optisam-backend/catalog-service/Catalog-TestData-import.xlsx/Catalog-TestData-import.xlsx', filename: 'Catalog-TestData-import.xlsx', contentType: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' }

        # * header Content-Type = 'multipart/form-data'
        # * multipart file file = { value: '#(file1_tmp)', filename: 'Catalog-TestData-import.xlsx', contentType: "application/xlsx" }
        # When method post
        # Then status 200 
