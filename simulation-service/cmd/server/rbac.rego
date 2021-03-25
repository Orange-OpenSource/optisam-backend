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

user_apis := {"/optisam.simulation.v1.SimulationService/GetConfigData","/optisam.simulation.v1.SimulationService/SimulationByMetric",
"/optisam.simulation.v1.SimulationService/SimulationByHardware","/optisam.simulation.v1.SimulationService/ListConfig"}