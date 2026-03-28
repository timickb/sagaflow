package utils

func IsStrNilOrEmpty(str *string) bool {
	if str == nil {
		return true
	}
	return *str == ""
}
