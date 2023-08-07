Feature: Simulation Service Test : Admin

Background:
     * def productServiceUrl = "https://optisam-product-dev.apps.fr01.paas.tech.orange"
      * url productServiceUrl+'/api/v1'
      * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
      * callonce read('common.feature') credentials
      * def access_token = response.access_token
      * header Authorization = 'Bearer '+access_token
      * def data = read('data.json')
      * def scope = 'API'


    Scenario: To verify the result when clicked on matric tab
        Given  path 'product/aggregations','editors'
        And params {scope:'#(scope)'}
        When method get
        Then status 200 
       # And match response.editor[*] contains ["Test3-3", "Dummy", "Microsoft", "Axicon Auto ID", "Redhat", "ADOBE-READER", "ABT", "Random","Klaxoon","ATNOS","Software AG","Bungie","Oracle","Adobe","AceBIT","001DemoEditor1","IBM"]
      
        
        
    Scenario: Choosing Editor for Metric simulation
        Given path 'product/editors','products' 
        And params {scopes:'#(scope)', editor:'#(data.SwidTag.editor2)'}
        When method get
        Then status 200
        And match response contains deep {"products":[{"swidTag": "Adobe_Reader_Adobe_7.0.12_build_2318", "name": "Adobe Reader","version": "7.0.12 build 2318"}]}

        




#-------------------------------------for aggregation------------------------------------------#
    Scenario: Verify the selection of Aggregtion Option
        Given path 'product','aggregations','editors'
        And params { scope:'#(scope)'}
        When method get 
        Then status 200
       # And match response.editor[*] contains ["Test3-3", "Dummy", "Microsoft", "Axicon Auto ID", "Redhat", "ADOBE-READER", "ABT", "Random","Klaxoon","ATNOS","Software AG","Bungie","Oracle","Adobe","AceBIT","001DemoEditor1","IBM"]



    Scenario: Verify the selection of Editor in  Aggregtion 
        Given path 'product/aggregations'
        And params { page_num:1,page_size:50,sort_by:'aggregation_name',sort_order:'asc',scope:'#(scope)',search_params.product_editor.filteringkey:'#(data.SwidTag.editor2)'}
        When method get 
        Then status 200
        
         

    Scenario: To verify the Editor selection for cost simulation
        Given path 'product','simulation','Adobe','rights'
        And params { scope:'#(scope)',editor:'#(data.SwidTag.editor2)'}
        When method get 
        Then status 200 
        And match response.editor_rights[*].metric_name contains ["instance_1"]
    
