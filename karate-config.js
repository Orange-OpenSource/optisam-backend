var envSuperAdminUserName= java.lang.System.getenv('APP_SUPER_ADMIN_USERNAME');
var envSuperAdminUserPassword= java.lang.System.getenv('APP_SUPER_ADMIN_PASSWORD');
var envAdminuserName= java.lang.System.getenv('APP_ADMIN_USERNAME');
var envAdminuserPassword= java.lang.System.getenv('APP_ADMIN_PASSWORD');
var envuserName= java.lang.System.getenv('APP_USER_USERNAME');
var envuserPassword= java.lang.System.getenv('APP_USER_PASSWORD');

//require('dotenv').config()
//console.log(process.env)
function fn() {
  var env = karate.env; // get system property 'karate.env'

  if (!env) {
    env = 'dev'; //dev,int

  }
  karate.log('karate.env system property:', env);

  //Configure Karate
  karate.configure('logPrettyRequest', true)
  karate.configure('logPrettyResponse', true)
  karate.configure('ssl', true)
  // karate.configure('ssl', { trustAll: true });
  karate.configure('abortedStepsShouldPass', true)
  

  // karate.configure('connectTimeout', 5000);
  // karate.configure('readTimeout', 5000);

  //Config variables
  var config = {
    env: env,
    access_token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiJyZWcudXNlckB0ZXN0LmNvbSIsIkxvY2FsZSI6ImVuIiwiUm9sZSI6IkFkbWluIiwiU29jcGVzIjpbIk9TUCIsIlRTVCJdLCJleHAiOjE2MDY5ODMzMTcsImlhdCI6MTYwNjk3NjExNywiaXNzIjoiT3JhbmdlIiwic3ViIjoiQWNjZXNzIFRva2VuIn0.OfxXAQP1EU-9sWhqCj_0My3hp4lxnNxyUk3FjjyzQMOArbM3Y6yjb1lKWSZoK6TlWJbzZRaritzBDhVBOGb0BLXmRCIRXCywqojrdNtSXe_cBKe1Ohrfdg4V-8Vy5Ip4cW5wg8Yx7-Bn40Wl_39X3a8XOkTUWvcbDsZ8uPUsDQ56h4-VUNNT6FMF0zmo4HE45MdZITGezUoth2dn7b6I9TC49RtgKkuXWJ5BiB5zio2aRFpZDAYKc9BC4MWfbfuTG6qJ3RcMBm1yW5pbtQMy03QM7OJXG8ZzLE2E5fzXwFyTXRUzbtRKE9RZhrJSpYn1jjJS6CGAWwEYqU8v2C0wrA",
    authServiceUrl: "",
    //AdminAccount_UserName: "",
    //AdminAccount_Password: "",
    //UserAccount_Username: "",
    //UserAccount_password: ""
    AdminAccount_UserName: envSuperAdminUserName,
    AdminAccount_Password: envSuperAdminUserPassword,
    UserAccount_Username: envuserName,
    UserAccount_password: envuserPassword
  };

  if (env == 'local') {
    config.authServiceUrl = "https://optisam-auth-int.apps.fr01.paas.tech.orange" 
    config.applicationServiceUrl = "http://localhost:7090"
    config.productServiceUrl = "http://localhost:12091"
    config.dpsServiceUrl = "http://localhost:10001"
    config.simulationServiceUrl = "http://localhost:22091"
    config.importServiceUrl = "http://localhost:9092"
   } else if (env == 'performance') {
    config.authServiceUrl = "https://optisam-auth-performance.apps.fr01.paas.tech.orange"
    config.accountServiceUrl = "https://optisam-account-performance.apps.fr01.paas.tech.orange"
    config.applicationServiceUrl = "https://optisam-application-performance.apps.fr01.paas.tech.orange"
    config.productServiceUrl = "https://optisam-product-performance.apps.fr01.paas.tech.orange"
    config.dpsServiceUrl = "https://optisam-dps-performance.apps.fr01.paas.tech.orange"
    config.importServiceUrl = "https://optisam-import-performance.apps.fr01.paas.tech.orange"
    config.equipmentServiceUrl = "https://optisam-equipment-performance.apps.fr01.paas.tech.orange"
    config.licenseServiceUrl = "https://optisam-license-performance.apps.fr01.paas.tech.orange"
    config.reportServiceUrl = "https://optisam-report-performance.apps.fr01.paas.tech.orange"
    config.metricServiceUrl = "https://optisam-metric-performance.apps.fr01.paas.tech.orange"
    config.simulationServiceUrl = "https://optisam-simulation-performance.apps.fr01.paas.tech.orange"
  } else if (env == 'dev') {
    config.authServiceUrl = "https://optisam-auth-dev.apps.fr01.paas.tech.orange" 
    config.accountServiceUrl = "https://optisam-account-dev.apps.fr01.paas.tech.orange"
    config.applicationServiceUrl = "https://optisam-application-dev.apps.fr01.paas.tech.orange"
    config.productServiceUrl = "https://optisam-product-dev.apps.fr01.paas.tech.orange"
    config.dpsServiceUrl = "https://optisam-dps-dev.apps.fr01.paas.tech.orange"
    config.importServiceUrl = "https://optisam-import-dev.apps.fr01.paas.tech.orange"
    config.equipmentServiceUrl = "https://optisam-equipment-dev.apps.fr01.paas.tech.orange"
    config.licenseServiceUrl = "https://optisam-license-dev.apps.fr01.paas.tech.orange"
    config.reportServiceUrl = "https://optisam-report-dev.apps.fr01.paas.tech.orange"
    config.metricServiceUrl = "https://optisam-metric-dev.apps.fr01.paas.tech.orange"
    config.simulationServiceUrl = "https://optisam-simulation-dev.apps.fr01.paas.tech.orange"
  } else if (env == 'int') {
    config.authServiceUrl = "https://optisam-auth-int.apps.fr01.paas.tech.orange" 
    config.accountServiceUrl = "https://optisam-account-int.apps.fr01.paas.tech.orange"
    config.applicationServiceUrl = "https://optisam-application-int.apps.fr01.paas.tech.orange"
    config.productServiceUrl = "https://optisam-product-int.apps.fr01.paas.tech.orange"
    config.dpsServiceUrl = "https://optisam-dps-int.apps.fr01.paas.tech.orange"
    config.importServiceUrl = "https://optisam-import-int.apps.fr01.paas.tech.orange"
    config.equipmentServiceUrl = "https://optisam-equipment-int.apps.fr01.paas.tech.orange"
    config.licenseServiceUrl = "https://optisam-license-int.apps.fr01.paas.tech.orange"
    config.reportServiceUrl = "https://optisam-report-int.apps.fr01.paas.tech.orange"
    config.metricServiceUrl = "https://optisam-metric-int.apps.fr01.paas.tech.orange"
    config.simulationServiceUrl = "https://optisam-simulation-int.apps.fr01.paas.tech.orange"
  }
  return config;
}
