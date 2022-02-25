package rbac

default allow = true

# Allow admins to do anything.
# allow {
# 	# roles["Admin"][input.role]
# }

# Normal Users
allow {
  user_apis[input.api]
  input.role = "User"
}

# roles := {"Admin":{"SuperAdmin","Admin"},"Normal":{"User"}}

user_apis := {"/optisam.license.v1.LicenseService/ListAcqRightsForProductAggregation","/optisam.license.v1.LicenseService/ListAcqRightsForApplicationsProduct","/optisam.license.v1.LicenseService/ProductLicensesForMetric","/optisam.license.v1.LicenseService/LicensesForEquipAndMetric","/optisam.license.v1.LicenseService/ListAcqRightsForProduct"}