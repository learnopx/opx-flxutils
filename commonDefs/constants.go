package commonDefs

//L2 types
const (
	IfTypePort = iota
	IfTypeLag
	IfTypeVlan
	IfTypeP2P
	IfTypeBcast
	IfTypeLoopback
	IfTypeSecondary
	IfTypeVirtual
	IfTypeVtep
	IfTypeNull
)

func GetIfTypeName(ifType int) string {
	switch ifType {
	case IfTypePort:
		return "Port"
	case IfTypeLag:
		return "Lag"
	case IfTypeVlan:
		return "Vlan"
	case IfTypeVtep:
		return "Vtep"
	default:
		return "Unknown"
	}
}
