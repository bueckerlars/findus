package domain

type RegistrationMode string

const (
	RegistrationAdminOnly RegistrationMode = "admin_only"
	RegistrationInvite    RegistrationMode = "invite"
	RegistrationOpen      RegistrationMode = "open"
)

func ParseRegistrationMode(s string) (RegistrationMode, bool) {
	switch RegistrationMode(s) {
	case RegistrationAdminOnly, RegistrationInvite, RegistrationOpen:
		return RegistrationMode(s), true
	default:
		return "", false
	}
}

const SettingRegistrationMode = "registration_mode"
