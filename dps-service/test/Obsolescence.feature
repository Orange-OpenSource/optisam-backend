Feature: oblolescence Test 

Background:
    * url applicationServiceUrl +'/api/v1/application'
   
    * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
    * callonce read('common.feature') credentials
    * def access_token = response.access_token
    * header Authorization = 'Bearer '+access_token
    * def data = read('data.json')
    * def scope = 'API' 

Scenario: To get the obsolescence Domain 
    Given path 'obsolescence/domains'
    And params {scope :'#(scope)'}
    When method get 
    Then status 200 
 
Scenario: To get the risk metrix
    Given path 'obsolescence/matrix'
    And params { scope :'#(scope)'}
    When method get 
    Then status 200
    And match response.risk_matrix[*].domain_critic_name contains ["Neutral"]

Scenario: To get the Response of Domain cricity
    Given path 'obsolescence/meta/domaincriticity'
    When method get 
    Then status 200
    And match response.domain_criticity_meta[*].domain_critic_name contains ["Non Critical"]

 Scenario: To get the Response of Domain cricity
    Given path 'obsolescence/meta/maintenancecriticity'
    When method get 
    Then status 200   
And match response.maintenance_criticity_meta[*].maintenance_critic_name contains ["Level 3"]

Scenario: To get the Response of maintenance
    Given path 'obsolescence/maintenance'
    And params { scope :'#(scope)'}
    When method get 
    Then status 200
    And match response.maintenance_criticy[*].maintenance_critic_id contains [3117]

    Scenario: To get the Response of Domain
    Given path 'domains'
    And params { scope :'#(scope)'}
    When method get 
    Then status 200 
    And match response.domains[*] contains ["internet"]

Scenario: To get the Response of Risk 
    Given path 'obsolescence/meta/risks'
    When method get 
    Then status 200
    And match response.risk_meta[*].risk_name contains ["Low"]


    Scenario: To get the Response of Matrix
    Given path 'obsolescence/matrix'
    And params { scope:'#(scope)'}
    When method get 
    Then status 200
    And match response.risk_matrix[*].domain_critic_name contains ["Neutral"]


    Scenario: To modify the Domain Criticality
        Given path 'obsolescence/domains'
        And request data.Modify_Criticity
        When method post 
        Then status 500 



    Scenario: To  Modify time criticity 
        Given path 'obsolescence/maintenance'
        And request  data.Time_Criticity
        When method post
        Then status 200

    Scenario: To modify the risk matrix 
        Given path 'obsolescence/matrix'
        And request data.Risk_Metrix
        When method post 
        Then status 200 
        
        
   

