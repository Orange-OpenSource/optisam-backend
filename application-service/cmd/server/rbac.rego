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
user_apis := {"/optisam.applications.v1.ApplicationService/ListApplications","/optisam.applications.v1.ApplicationService/ListInstances"}