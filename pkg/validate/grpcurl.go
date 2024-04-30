package validate

import (
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
)

func validateGrpcUrl(field validator.FieldLevel) bool {
	u, err := url.Parse(field.Field().String())
	if err != nil {
		return false
	}

	// Check the scheme
	switch u.Scheme {
	case "dns", "passthrough":
		if !strings.Contains(u.Path, ":") {
			return false
		}
	case "ipv4":
		ipPort := strings.SplitN(u.Host[1:], ":", 2)
		ip := net.ParseIP(ipPort[0])
		if ip == nil || ip.To4() == nil {
			return false
		}
	case "ipv6":
		ipPort := strings.SplitN(u.Host[2:], "]:", 2)
		ip := net.ParseIP(ipPort[0])
		if ip == nil || ip.To4() != nil {
			return false
		}
	case "unix", "unix-abstract":
		_, err := os.Stat(u.Path)
		if err != nil {
			return false
		}
	default:
		return false
	}

	return true
}
