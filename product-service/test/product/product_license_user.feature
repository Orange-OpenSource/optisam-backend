@product
Feature: Product Service Test - Normal user

  Background:
    * url licenseServiceUrl+'/api/v1'
    * def credentials = {username:#(UserAccount_Username), password:#(UserAccount_password)}
    * callonce read('../common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

@get
Scenario: To verify Computed licenses for the product Adobe_Media_Server_Adobe
  Given path 'license/product/'+data.get_productDetail.swidTag+'/acquiredrights'
  And params {scope:'#(scope)'}
  When method get
  Then status 200
* def result = karate.jsonPath(response, "$.acq_rights[?(@.numCptLicences=='"+data.get_productDetail.numCptLicences+"')]")[0]
* match result == data.get_productDetail
* match response.acq_rights[0].numCptLicences == data.get_productDetail.numCptLicences

@get
Scenario: To verify Acquired Licenses for the product Adobe_Media_Server_Adobe
  Given path 'license/product/'+data.get_productDetail.swidTag+'/acquiredrights'
  And params {scope:'#(scope)'}
  When method get
  Then status 200
* match response.acq_rights[0].numAcqLicences == data.get_productDetail.numAcqLicences

@get
Scenario: To verify Delta (licenses) for the product Adobe_Media_Server_Adobe
  Given path 'license/product/'+data.get_productDetail.swidTag+'/acquiredrights'
  And params {scope:'#(scope)'}
  When method get
  Then status 200
* match response.acq_rights[0].deltaNumber == data.get_productDetail.deltaNumber

@get
Scenario: To verify Delta Cost for the product Adobe_Media_Server_Adobe
  Given path 'license/product/'+data.get_productDetail.swidTag+'/acquiredrights'
  And params {scope:'#(scope)'}
  When method get
  Then status 200
* match response.acq_rights[0].deltaCost == data.get_productDetail.deltaCost


@get
Scenario: To verify Total Cost for the product Adobe_Media_Server_Adobe
  Given path 'license/product/'+data.get_productDetail.swidTag+'/acquiredrights'
  And params {scope:'#(scope)'}
  When method get
  Then status 200
* match response.acq_rights[0].totalCost == data.get_productDetail.totalCost

@get
Scenario: To verify SKU for the product Adobe_Media_Server_Adobe
  Given path 'license/product/'+data.get_productDetail.swidTag+'/acquiredrights'
  And params {scope:'#(scope)'}
  When method get
  Then status 200
* match response.acq_rights[0].SKU == data.get_productDetail.SKU

Scenario: To verify aggregationName for Aggregation Type of product Oracle_enterprise_database
  Given path 'license/aggregation/'+data.get_Aggregation_product_Details.aggregationName+'/acquiredrights'
  And params {scope:'#(scope)'}
  When method get
  Then status 200
  And match response.acq_rights[0].aggregationName == data.get_Aggregation_product_Details.aggregationName

Scenario: To verify Computed licenses for aggregated product Oracle_enterprise_database
  Given path 'license/aggregation/'+data.get_Aggregation_product_Details.aggregationName+'/acquiredrights'
  And params {scope:'#(scope)'}
  When method get
  Then status 200
  And match response.acq_rights[0].numCptLicences == data.get_Aggregation_product_Details.numCptLicences

Scenario: To verify Acquired Licenses for aggregated product Oracle_enterprise_database
  Given path 'license/aggregation/'+data.get_Aggregation_product_Details.aggregationName+'/acquiredrights'
  And params {scope:#(scope)}
  When method get
  Then status 200
  And response.acq_rights[0].numAcqLicences == data.get_Aggregation_product_Details.numAcqLicences

