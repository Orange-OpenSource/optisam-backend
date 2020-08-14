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
user_apis := {"/v1.AcqRightsService/ListAcqRights","/v1.AcqRightsService/ListAcqRightsAggregation",
"/v1.AcqRightsService/ListAcqRightsAggregationRecords","/v1.AcqRightsService/ListAcqRightsEditors","/v1.AcqRightsService/ListAcqRightsMetrics"}