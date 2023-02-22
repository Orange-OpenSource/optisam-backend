package rbac

default allow = false

# Allow admins to do anything.
allow {
	roles["Admin"][input.role]
}

# Normal Users
allow {
  user_apis[input.api]
  input.role = "User"
}


roles := {"Admin":{"SuperAdmin","Admin"},"Normal":{"User"}}

user_apis := {"/optisam.dps.v1.DpsService/DashboardQualityOverview","/optisam.dps.v1.DpsService/DashboardDataFailureRate","/optisam.dps.v1.DpsService/ListFailureReasonsRatio"}
