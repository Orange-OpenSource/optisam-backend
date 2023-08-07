@license
Feature: License Service Test

  Background:
    * url licenseServiceUrl+'/api/v1/license'
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API'


  Scenario: Validate Licence for instance_1 for product Adobe
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
   * match response.acq_rights[0].computedCost == data.sag_processor.computedCost
   * match response.acq_rights[0].purchaseCost == data.sag_processor.purchaseCost
   * match response.acq_rights[0].avgUnitPrice == data.sag_processor.avgUnitPrice

  Scenario: Validate Licence for instance_1 for product Redhat Enterprise Linux Server with mantance cost
    Given path 'product', data.instance_1_Redhat_Enterprise.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200
    * match response.acq_rights[0].numCptLicences == data.instance_1_Redhat_Enterprise.Computed_licenses
    * match response.acq_rights[0].numAcqLicences == data.instance_1_Redhat_Enterprise.Acquired_Licenses
    * match response.acq_rights[0].deltaNumber == data.instance_1_Redhat_Enterprise.Delta_licenses
    * match response.acq_rights[0].deltaCost == data.instance_1_Redhat_Enterprise.Delta_Cost
    * match response.acq_rights[0].totalCost == data.instance_1_Redhat_Enterprise.Total_Cost
    * match response.acq_rights[0].metric == data.instance_1_Redhat_Enterprise.metric
    * match response.acq_rights[0].computedCost == data.instance_1_Redhat_Enterprise.computedCost
    * match response.acq_rights[0].purchaseCost == data.instance_1_Redhat_Enterprise.purchaseCost
    * match response.acq_rights[0].avgUnitPrice == data.instance_1_Redhat_Enterprise.avgUnitPrice

##  Scenario: Validate Licence for static_1 for product Random_product
 #   Given path 'product', data.static_standard.swidTag, 'acquiredrights'
 #   And params {scope: '#(scope)'}
 #   When method get
 #   Then status 200
 #   * match response.acq_rights[0].numCptLicences == data.static_standard.Computed_licenses
 #   * match response.acq_rights[0].numAcqLicences == data.static_standard.Acquired_Licenses
 #   * match response.acq_rights[0].deltaNumber == data.static_standard.Delta_licenses
 #   * match response.acq_rights[0].deltaCost == data.static_standard.Delta_Cost
 #   * match response.acq_rights[0].totalCost == data.static_standard.Total_Cost
 #   * match response.acq_rights[0].metric == data.static_standard.metric
 #   * match response.acq_rights[0].computedCost == data.static_standard.computedCost
 #   * match response.acq_rights[0].purchaseCost == data.static_standard.purchaseCost
 #   * match response.acq_rights[0].avgUnitPrice == data.static_standard.avgUnitPrice

  Scenario: Validate Licence for 8TB,instance_62 for product Red Hat Ceph Storage 15
    Given path 'product', data.EightTB_instance_62.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.EightTB_instance_62.SKU+"')]")[0]

    * match result.numCptLicences == data.EightTB_instance_62.Computed_licenses
    * match result.numAcqLicences == data.EightTB_instance_62.Acquired_Licenses
    * match result.deltaNumber == data.EightTB_instance_62.Delta_licenses
    * match result.deltaCost == data.EightTB_instance_62.Delta_Cost
    * match result.totalCost == data.EightTB_instance_62.Total_Cost
    * match result.metric == data.EightTB_instance_62.metric
    * match result.computedCost == data.EightTB_instance_62.computedCost
    * match result.purchaseCost == data.EightTB_instance_62.purchaseCost
    * match result.avgUnitPrice == data.EightTB_instance_62.avgUnitPrice


  Scenario: Validate Licence for 8TB for product Red Hat Ceph Storage 15
    Given path 'product', data.EightTB.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200  

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.EightTB.SKU+"')]")[0]
    * match result.numCptLicences == data.EightTB.Computed_licenses
    * match result.numAcqLicences == data.EightTB.Acquired_Licenses
    * match result.deltaNumber == data.EightTB.Delta_licenses
    * match result.deltaCost == data.EightTB.Delta_Cost
    * match result.totalCost == data.EightTB.Total_Cost
    * match result.metric == data.EightTB.metric
    * match result.computedCost == data.EightTB.computedCost
    * match result.purchaseCost == data.EightTB.purchaseCost
    * match result.avgUnitPrice == data.EightTB.avgUnitPrice

  Scenario: Validate Licence for instance_1 for product Redhat Openshift Platform with coef 1
    Given path 'product', data.instance_1_coef_1.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200
    
  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.instance_1_coef_1.SKU+"')]")[0]

    * match result.numCptLicences == data.instance_1_coef_1.Computed_licenses
    * match result.numAcqLicences == data.instance_1_coef_1.Acquired_Licenses
    * match result.deltaNumber == data.instance_1_coef_1.Delta_licenses
    * match result.deltaCost == data.instance_1_coef_1.Delta_Cost
    * match result.totalCost == data.instance_1_coef_1.Total_Cost
    * match result.metric == data.instance_1_coef_1.metric
    * match result.computedCost == data.instance_1_coef_1.computedCost
    * match result.purchaseCost == data.instance_1_coef_1.purchaseCost
    * match result.avgUnitPrice == data.instance_1_coef_1.avgUnitPrice
    

  Scenario: Validate Licence for hyperthreading for product Product_Hyperthreading
    Given path 'product', data.hyperthreading.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.hyperthreading.SKU+"')]")[0]

    * match result.numCptLicences == data.hyperthreading.Computed_licenses
    * match result.numAcqLicences == data.hyperthreading.Acquired_Licenses
    * match result.deltaNumber == data.hyperthreading.Delta_licenses
    * match result.deltaCost == data.hyperthreading.Delta_Cost
    * match result.totalCost == data.hyperthreading.Total_Cost
    * match result.metric == data.hyperthreading.metric
    * match result.computedCost == data.hyperthreading.computedCost
    * match result.purchaseCost == data.hyperthreading.purchaseCost
    * match result.avgUnitPrice == data.hyperthreading.avgUnitPrice

  Scenario: Validate Licence for instance_12,memory_256TB  for product Red Hat Ceph Storage
   
    Given path 'product', data.instance_12_memory_256TB.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200
  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.instance_12_memory_256TB.SKU+"')]")[0]

    * match result.numCptLicences == data.instance_12_memory_256TB.Computed_licenses
    * match result.numAcqLicences == data.instance_12_memory_256TB.Acquired_Licenses
    * match result.deltaNumber == data.instance_12_memory_256TB.Delta_licenses
    * match result.deltaCost == data.instance_12_memory_256TB.Delta_Cost
    * match result.totalCost == data.instance_12_memory_256TB.Total_Cost
    * match result.metric == data.instance_12_memory_256TB.metric
    * match result.computedCost == data.instance_12_memory_256TB.computedCost
    * match result.purchaseCost == data.instance_12_memory_256TB.purchaseCost
    * match result.avgUnitPrice == data.instance_12_memory_256TB.avgUnitPrice

  Scenario: Validate Licence for instance_50,memory_1000TB  for product Red Hat Ceph Storage
    Given path 'product', data.instance_50_memory_1000TB.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200
  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.instance_50_memory_1000TB.SKU+"')]")[0]

    * match result.numCptLicences == data.instance_50_memory_1000TB.Computed_licenses
    * match result.numAcqLicences == data.instance_50_memory_1000TB.Acquired_Licenses
    * match result.deltaNumber == data.instance_50_memory_1000TB.Delta_licenses
    * match result.deltaCost == data.instance_50_memory_1000TB.Delta_Cost
    * match result.totalCost == data.instance_50_memory_1000TB.Total_Cost
    * match result.metric == data.instance_50_memory_1000TB.metric
    * match result.computedCost == data.instance_50_memory_1000TB.computedCost
    * match result.purchaseCost == data.instance_50_memory_1000TB.purchaseCost
    * match result.avgUnitPrice == data.instance_50_memory_1000TB.avgUnitPrice


  Scenario: Validate Licence for User for product Random product 1
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
    * match result.metric == data.User_License.metric
    * match result.computedCost == data.User_License.computedCost
    * match result.purchaseCost == data.User_License.purchaseCost
    * match result.avgUnitPrice == data.User_License.avgUnitPrice


  Scenario: Validate Licence for ibm_pvu for product IBM Cognos Analytics
    Given path 'product', data.IBM_PUV_IBM_Cognos_Analytics.swidTag, 'acquiredrights'
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




  Scenario: Validate Licence for oracle.processor for product Oracle Weblogic Server
    Given path 'product', data.Oracle_Weblogic_Server_Oracle_Weblogic.swidTag, 'acquiredrights'
    And params {scope: '#(scope)'}
    When method get
    Then status 200
  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.Oracle_Weblogic_Server_Oracle_Weblogic.SKU+"')]")[0]
    

    * match result.numCptLicences == data.Oracle_Weblogic_Server_Oracle_Weblogic.Computed_licenses
    * match result.numAcqLicences == data.Oracle_Weblogic_Server_Oracle_Weblogic.Acquired_Licenses
    * match result.deltaNumber == data.Oracle_Weblogic_Server_Oracle_Weblogic.Delta_licenses
    * match result.deltaCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.Delta_Cost
    * match result.totalCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.Total_Cost
    * match result.metric == data.Oracle_Weblogic_Server_Oracle_Weblogic.metric
    * match result.computedCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.computedCost
    * match result.purchaseCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.purchaseCost
    * match result.avgUnitPrice == data.Oracle_Weblogic_Server_Oracle_Weblogic.avgUnitPrice



  Scenario: Validate license for Application Type carala for product Adobe Media Server
    Given path 'applications/car/products', data.instance_1.swidTag
    And params {scope: '#(scope)'}
    When method get
    Then status 200
    
    * match response.acq_rights[0].numCptLicences == data.instance_1.Computed_licenses
    * match response.acq_rights[0].numAcqLicences == data.instance_1.Acquired_Licenses
    * match response.acq_rights[0].deltaNumber == data.instance_1.Delta_licenses
    * match response.acq_rights[0].deltaCost == data.instance_1.Delta_Cost
    * match response.acq_rights[0].totalCost == data.instance_1.Total_Cost
    * match response.acq_rights[0].metric == data.instance_1.metric

  Scenario: Validate Licence for Application Type carala for product IBM Cognos Analytics
    Given path 'applications/car/products', data.IBM_PUV_IBM_Cognos_Analytics.swidTag
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.IBM_PUV_IBM_Cognos_Analytics.SKU+"')]")[0]
    
    * match result.numCptLicences == data.IBM_PUV_IBM_Cognos_Analytics.Computed_licenses
    * match result.numAcqLicences == data.IBM_PUV_IBM_Cognos_Analytics.Acquired_Licenses
    * match result.deltaNumber == data.IBM_PUV_IBM_Cognos_Analytics.Delta_licenses
    * match result.deltaCost == data.IBM_PUV_IBM_Cognos_Analytics.Delta_Cost
    * match result.totalCost == data.IBM_PUV_IBM_Cognos_Analytics.Total_Cost


  Scenario: Validate Licence for Application Type fabra for product IBM Cognos Analytics
    Given url 'https://optisam-license-pc.apps.fr01.paas.tech.orange/api/v1/license'
    And path 'applications/fab/products', data.Oracle_Weblogic_Server_Oracle_Weblogic.swidTag
    And params {scope: '#(scope)'}
    When method get
    Then status 200

  * def result = karate.jsonPath(response, "$.acq_rights[?(@.SKU=='"+data.Oracle_Weblogic_Server_Oracle_Weblogic.SKU+"')]")[0]
    
    * match result.numCptLicences == data.Oracle_Weblogic_Server_Oracle_Weblogic.Computed_licenses
    * match result.numAcqLicences == data.Oracle_Weblogic_Server_Oracle_Weblogic.Acquired_Licenses
    * match result.deltaNumber == data.Oracle_Weblogic_Server_Oracle_Weblogic.Delta_licenses
    * match result.deltaCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.Delta_Cost
    * match result.totalCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.Total_Cost
    * match result.metric == data.Oracle_Weblogic_Server_Oracle_Weblogic.metric
    * match result.computedCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.computedCost
    * match result.purchaseCost == data.Oracle_Weblogic_Server_Oracle_Weblogic.purchaseCost
    * match result.avgUnitPrice == data.Oracle_Weblogic_Server_Oracle_Weblogic.avgUnitPrice