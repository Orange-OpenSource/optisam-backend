@license
Feature: License Service Test - Compliance for Metrics inm,acs,sag,pvu : admin

  Background:
  # * def licenseServiceUrl = "https://optisam-license-int.apps.fr01.paas.tech.orange"
    * url licenseServiceUrl+'/api/v1/license'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'

    Scenario: Validate licence for sag.processor for product Software AG WebMethodsar
      Given path 'product', data.sag_processor.swidTag, 'acquiredrights'
      And params {scope: '#(scope)'}
      When method get
      Then status 200
     * match response.acq_rights[0].metric == data.sag_processor.metric
     * match response.acq_rights[0].numCptLicences == data.sag_processor.Computed_licenses
     * match response.acq_rights[0].numAcqLicences == data.sag_processor.Acquired_Licenses
     * match response.acq_rights[0].deltaNumber == data.sag_processor.Delta_licenses
     * match response.acq_rights[0].deltaCost == data.sag_processor.Delta_Cost
     * match response.acq_rights[0].totalCost == data.sag_processor.Total_Cost

    Scenario: Validate Licence for OS instance like instance_1 for product Adobe
      Given path 'product', data.instance_1.swidTag, 'acquiredrights'
      And params {scope: '#(scope)'}
      When method get
      Then status 200
      
     * match response.acq_rights[0].numCptLicences == data.instance_1.Computed_licenses
      * match response.acq_rights[0].numAcqLicences == data.instance_1.Acquired_Licenses
      * match response.acq_rights[0].deltaNumber == data.instance_1.Delta_licenses
      * match response.acq_rights[0].deltaCost == data.instance_1.Delta_Cost
      * match response.acq_rights[0].totalCost == data.instance_1.Total_Cost
      * match response.acq_rights[0].metric == data.instance_1.metric


    Scenario: Validate Licence for static_1 for product Random_product
      Given path 'product', data.static_standard.swidTag, 'acquiredrights'
      And params {scope: '#(scope)'}
      When method get
      Then status 200
  
      * match response.acq_rights[0].numCptLicences == data.static_standard.Computed_licenses
      * match response.acq_rights[0].numAcqLicences == data.static_standard.Acquired_Licenses
      * match response.acq_rights[0].deltaNumber == data.static_standard.Delta_licenses
      * match response.acq_rights[0].deltaCost == data.static_standard.Delta_Cost
      * match response.acq_rights[0].totalCost == data.static_standard.Total_Cost

    Scenario: Validate Licence for User for product Random product 2
      Given path 'product', data.User_License.swidTag, 'acquiredrights'
      And params {scope: '#(scope)'}
      When method get
      Then status 200
    * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.User_License.SKU+"')]")[0]
  
      * match result.numCptLicences == data.User_License.Computed_licenses
      * match result.numAcqLicences == data.User_License.Acquired_Licenses
      * match result.deltaNumber == data.User_License.Delta_licenses
      * match result.deltaCost == data.User_License.Delta_Cost
      * match result.totalCost == data.User_License.Total_Cost    

    Scenario: Validate Licence for hyperthreading for product Product_Hyperthreading
      Given path 'product', data.hyperthreading.swidTag, 'acquiredrights'
      And params {scope: '#(scope)'}
      When method get
      Then status 200
  
    * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.hyperthreading.SKU+"')]")[0]
  
      * match result.numCptLicences == data.hyperthreading.Computed_licenses
      * match result.numAcqLicences == data.hyperthreading.Acquired_Licenses
      * match result.deltaNumber == data.hyperthreading.Delta_licenses

    Scenario: Validate Licence for 8TB for product Red Hat Ceph Storage 15
      Given path 'product', data.EightTB.swidTag, 'acquiredrights'
      And params {scope: '#(scope)'}
      When method get
      Then status 200  
  
    * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.EightTB.SKU+"')]")[0]
      * match result.numCptLicences == data.EightTB.Computed_licenses
      * match result.numAcqLicences == data.EightTB.Acquired_Licenses
      * match result.deltaNumber == data.EightTB.Delta_licenses



    Scenario: Validate Licence for ibm_pvu for product IBM Cognos Analytics
      Given url 'https://optisam-license-pc.apps.fr01.paas.tech.orange/api/v1/license'
      And path 'product', data.IBM_PUV_IBM_Cognos_Analytics.swidTag, 'acquiredrights'
      And params {scope: '#(scope)'}
      When method get
      Then status 200
    * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.IBM_PUV_IBM_Cognos_Analytics.SKU+"')]")[0]
      
      * match result.numCptLicences == data.IBM_PUV_IBM_Cognos_Analytics.Computed_licenses
      * match result.numAcqLicences == data.IBM_PUV_IBM_Cognos_Analytics.Acquired_Licenses
      * match result.deltaNumber == data.IBM_PUV_IBM_Cognos_Analytics.Delta_licenses
      * match result.deltaCost == data.IBM_PUV_IBM_Cognos_Analytics.Delta_Cost
      * match result.totalCost == data.IBM_PUV_IBM_Cognos_Analytics.Total_Cost
      * match result.metric == data.IBM_PUV_IBM_Cognos_Analytics.metric
      * match result.computedCost == data.IBM_PUV_IBM_Cognos_Analytics.computedCost
      * match result.purchaseCost == data.IBM_PUV_IBM_Cognos_Analytics.purchaseCost
      * match result.avgUnitPrice == data.IBM_PUV_IBM_Cognos_Analytics.avgUnitPrice
  
  
    Scenario: Validate Licence for ibm_pvu for product IBM DB2
      Given path 'product', data.IBM_PUV_IBM_DB2.swidTag, 'acquiredrights'
      And params {scope: '#(scope)'}
      When method get
      Then status 200
    * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.IBM_PUV_IBM_DB2.SKU+"')]")[0]
      
      * match result.numCptLicences == data.IBM_PUV_IBM_DB2.Computed_licenses
      * match result.numAcqLicences == data.IBM_PUV_IBM_DB2.Acquired_Licenses
      * match result.deltaNumber == data.IBM_PUV_IBM_DB2.Delta_licenses
      * match result.deltaCost == data.IBM_PUV_IBM_DB2.Delta_Cost
      * match result.totalCost == data.IBM_PUV_IBM_DB2.Total_Cost
      * match result.metric == data.IBM_PUV_IBM_DB2.metric
      * match result.computedCost == data.IBM_PUV_IBM_DB2.computedCost
      * match result.purchaseCost == data.IBM_PUV_IBM_DB2.purchaseCost
      * match result.avgUnitPrice == data.IBM_PUV_IBM_DB2.avgUnitPrice
  
    Scenario: Validate Licence for ibm_pvu for product IBM Websphere
      Given path 'product', data.IBM_PUV_IBM_Websphere.swidTag, 'acquiredrights'
      And params {scope: '#(scope)'}
      When method get
      Then status 200
    * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.IBM_PUV_IBM_Websphere.SKU+"')]")[0]
      
  
      * match result.numCptLicences == data.IBM_PUV_IBM_Websphere.Computed_licenses
      * match result.numAcqLicences == data.IBM_PUV_IBM_Websphere.Acquired_Licenses
      * match result.deltaNumber == data.IBM_PUV_IBM_Websphere.Delta_licenses
      * match result.deltaCost == data.IBM_PUV_IBM_Websphere.Delta_Cost
      * match result.totalCost == data.IBM_PUV_IBM_Websphere.Total_Cost
      * match result.metric == data.IBM_PUV_IBM_Websphere.metric
      * match result.computedCost == data.IBM_PUV_IBM_Websphere.computedCost
      * match result.purchaseCost == data.IBM_PUV_IBM_Websphere.purchaseCost
      * match result.avgUnitPrice == data.IBM_PUV_IBM_Websphere.avgUnitPrice