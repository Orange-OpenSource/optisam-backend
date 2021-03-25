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

user_apis := {"/v1.EquipmentService/EquipmentsTypes","/v1.EquipmentService/ListEquipments",
"/v1.EquipmentService/GetEquipment","/v1.EquipmentService/ListEquipmentParents",
"/v1.EquipmentService/ListEquipmentChildren","/v1.EquipmentService/ListEquipmentsForProductAggregation","/v1.EquipmentService/EquipmentsPerEquipmentType"}