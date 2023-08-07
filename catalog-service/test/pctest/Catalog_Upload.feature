@Upload
Feature: Catalog Upload Test

  Background:
    * url catalogServiceUrl
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')

Scenario: Schema validation for Upload
    Given path 'api/v1/catalog/bulkfileuploadlogs' 
    * def schema = data.Upload_Schema
    When method get
    Then status 200
    * response.uploadCatalogDataLogs == '#number? _ >= 0'
    * match response.uploadCatalogDataLogs == '#[_ > 0] schema'
    * match response.uploadCatalogDataLogs == '#[_ <= 50] schema'