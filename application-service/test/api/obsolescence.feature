@application
Feature: Application Service Test for Obsolescence

  Background:
  # * def applicationServiceUrl = "https://optisam-application-int.kermit-noprod-b.itn.intraorange"
    * url applicationServiceUrl+'/api/v1/application'
    * def credentials = {username:'admin@test.com', password: 'admin'}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'AUT'




  @get
  Scenario: get Obsolescence Domains
    Given path 'obsolescence/domains'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    * match response.domains_criticity == '#[]'

  @get
  Scenario: get Obsolescence Maintenance Criticity
    Given path 'obsolescence/maintenance'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    * match response.maintenance_criticy == '#[]'



     @create
  Scenario: Create Obsolescence Domian
    Given path 'obsolescence/domains'
    And request data.domains
    When method post
    Then status 200

      @create
  Scenario: Create Obsolescence maintenance
    Given path 'obsolescence/maintenance'
    And request data.maintenance
    When method post
    Then status 200

    @create
  Scenario: Create Obsolescence matrix
    Given path 'obsolescence/matrix'
    And request data.matrix
    When method post
    Then status 200
    

  @get
  Scenario: get Obsolescence Matrix
    Given path 'obsolescence/matrix'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    * match response.risk_matrix == '#[]'


     @get
  Scenario: get Obsolescence domaincriticity
    Given path 'obsolescence/meta/domaincriticity'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    * match response.domain_criticity_meta == '#[]'


     @get
  Scenario: get Obsolescence maintenancecriticity
    Given path 'obsolescence/meta/maintenancecriticity'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
     * match response.maintenance_criticity_meta == '#[]'


    @get
  Scenario: get Obsolescence risks
    Given path 'obsolescence/meta/risks'
    * params {scope:'#(scope)'}
    When method get
    Then status 200
    * match response.risk_meta == '#[]'