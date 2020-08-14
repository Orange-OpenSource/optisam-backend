package rbac

default allow = false

# Allow admins to do anything.
allow {
	roles["Admin"][input.role]
}

roles := {"Admin":{"SuperAdmin","Admin"},"Normal":{"User"}}
