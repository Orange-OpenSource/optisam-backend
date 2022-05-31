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
  karate.configure('abortedStepsShouldPass', true)

  // karate.configure('connectTimeout', 5000);
  // karate.configure('readTimeout', 5000);

  //Config variables
  var config = {
    env: env,
    access_token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiJyZWcudXNlckB0ZXN0LmNvbSIsIkxvY2FsZSI6ImVuIiwiUm9sZSI6IkFkbWluIiwiU29jcGVzIjpbIk9TUCIsIlRTVCJdLCJleHAiOjE2MDY5ODMzMTcsImlhdCI6MTYwNjk3NjExNywiaXNzIjoiT3JhbmdlIiwic3ViIjoiQWNjZXNzIFRva2VuIn0.OfxXAQP1EU-9sWhqCj_0My3hp4lxnNxyUk3FjjyzQMOArbM3Y6yjb1lKWSZoK6TlWJbzZRaritzBDhVBOGb0BLXmRCIRXCywqojrdNtSXe_cBKe1Ohrfdg4V-8Vy5Ip4cW5wg8Yx7-Bn40Wl_39X3a8XOkTUWvcbDsZ8uPUsDQ56h4-VUNNT6FMF0zmo4HE45MdZITGezUoth2dn7b6I9TC49RtgKkuXWJ5BiB5zio2aRFpZDAYKc9BC4MWfbfuTG6qJ3RcMBm1yW5pbtQMy03QM7OJXG8ZzLE2E5fzXwFyTXRUzbtRKE9RZhrJSpYn1jjJS6CGAWwEYqU8v2C0wrA",
    authServiceUrl: "" 
  };

  if (env == 'local') {
    config.authServiceUrl = "https://optisam-auth-int.apps.fr01.paas.tech.orange" 
    config.applicationServiceUrl = "http://localhost:7090"
    config.productServiceUrl = "http://localhost:12091"
    config.dpsServiceUrl = "http://localhost:10001"
    config.simulationServiceUrl = "http://localhost:22091"
    config.importServiceUrl = "http://localhost:9092"
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
