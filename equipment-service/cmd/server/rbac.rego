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

user_apis := {"/optisam.equipment.v1.EquipmentService/EquipmentsTypes","/optisam.equipment.v1.EquipmentService/ListEquipments",
"/optisam.equipment.v1.EquipmentService/GetEquipment","/optisam.equipment.v1.EquipmentService/ListEquipmentParents",
"/optisam.equipment.v1.EquipmentService/ListEquipmentChildren","/optisam.equipment.v1.EquipmentService/ListEquipmentsForProductAggregation","/optisam.equipment.v1.EquipmentService/EquipmentsPerEquipmentType"}