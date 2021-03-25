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

user_apis := {"/v1.LicenseService/ListAcqRightsForProductAggregation","/v1.LicenseService/ListAcqRightsForApplicationsProduct","/v1.LicenseService/ProductLicensesForMetric","/v1.LicenseService/LicensesForEquipAndMetric","/v1.LicenseService/ListAcqRightsForProduct"}