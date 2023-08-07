Feature: DPS Service Test - Data : admin user

Background:
  * url importServiceUrl+'/api/v1/import'

  * def credentials = {username:#(AdminAccount_UserName), password:#(AdminAccount_Password)}
  * callonce read('common.feature') credentials
  * def access_token = response.access_token
  * header Authorization = 'Bearer '+access_token
  * def data = read('data.json')
  * def scope = "API"



  Scenario:To verify the process to downlaod the error files 
    Given path 'download'
  And params {uploadId:#(data.Error_file.UploadId),downloadType:'#(data.Error_file.Type)',scope:'#(scope)'}
    When method get
    Then status 200 

Scenario:To verify the process to downlaod the  files 
    Given path 'download'
  And params {scope:'#(scope)',uploadId:#(data.Error_file.UploadId),downloadType:'#(data.Error_file.Type)'}
    When method get
    Then status 200 








