package rbac

default allow = false

# Allow admins to do anything.
allow {
	roles["Admin"][input.role]
}

# Allow users to do anything.
allow {
	roles["Normal"][input.role]
}

roles := {"Admin":{"SuperAdmin","Admin"},"Normal":{"User"}}