package report

func String(p Phase) string {
	switch p {
	case Phase_Unknown:
		return "Unknown"
	case Phase_Ready:
		return "Ready"
	case Phase_Running:
		return "Running"
	case Phase_Stopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}

func ToPhase(s string) Phase {
	switch s {
	case "Unknown":
		return Phase_Unknown
	case "Ready":
		return Phase_Ready
	case "Running":
		return Phase_Running
	case "Stopped":
		return Phase_Stopped
	default:
		return Phase_Unknown
	}
}
