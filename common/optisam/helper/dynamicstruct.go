package helper

func GetTypeInstance(datatype string) interface{} {
	switch datatype {
	case "string":
		return ""
	case "int":
		return 0
	case "float":
		return 0.0
	case "boolean":
		return true
	}
	return ""
}
