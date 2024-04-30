package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type CtxKey string

const CtxKeyAuthInfo CtxKey = "auth_info"

type AuthInfo struct {
	Subject   string
	Audiences []string

	ScopeEnabled bool
	Scopes       []string

	ResourceEnabled bool
	Resources       []string
}

func (auth *AuthInfo) FromClaims(claims jwt.MapClaims) {
	var ok bool
	if auth.Subject, ok = claims["sub"].(string); !ok {
		return
	}

	if auth.Audiences, ok = claims["aud"].(jwt.ClaimStrings); !ok {
		return
	}

	if scope, ok := claims["scope"].(string); ok {
		auth.Scopes = strings.Split(scope, " ")
	}

	if resource, ok := claims["resource"].(string); ok {
		auth.Resources = strings.Split(resource, " ")
	}
}

func (auth *AuthInfo) FromDNSNames(dnsNames []string) {
	for _, dnsName := range dnsNames {
		auth.FromDNSName(dnsName)
	}
}

func (auth *AuthInfo) FromDNSName(dnsName string) {
	names := strings.SplitN(dnsName, ".", 3)
	if len(names) < 2 {
		return
	}

	type_ := names[len(names)-1]
	var name string
	if len(names) == 3 {
		name = fmt.Sprintf("%s:%s", names[1], names[0])
	} else {
		name = names[0]
	}

	switch type_ {
	case "user":
		auth.Subject = name
	case "audience":
		auth.Audiences = append(auth.Audiences, name)
	case "scope":
		auth.Scopes = append(auth.Scopes, name)
	case "resource":
		auth.Resources = append(auth.Resources, name)
	}
}

func (auth *AuthInfo) Validate() error {
	if auth.Subject == "" {
		return errors.New("invalid subject")
	}

	return nil
}

func (auth *AuthInfo) ValidateAudience(audience string) bool {
	for _, a := range auth.Audiences {
		if a == audience {
			return true
		}
	}
	return false
}

func (auth *AuthInfo) ValidateScope(scope string) bool {
	if !auth.ScopeEnabled {
		return false
	}

	category := strings.SplitN(scope, ":", 2)[0]
	for _, s := range auth.Scopes {
		if s == category {
			return true
		}
		if s == scope {
			return true
		}
	}
	return false
}

func (auth *AuthInfo) ValidateScopes(scopes []string) ([]string, bool) {
	if !auth.ScopeEnabled {
		return scopes, false
	}

	var missing []string
	for _, scope := range scopes {
		if !auth.ValidateScope(scope) {
			missing = append(missing, scope)
		}
	}
	return missing, len(missing) == 0
}

func (auth *AuthInfo) ValidateResource(resource string) bool {
	if !auth.ResourceEnabled {
		return false
	}
	category := strings.SplitN(resource, ":", 2)[0]
	for _, r := range auth.Resources {
		if r == category {
			return true
		}
		if r == resource {
			return true
		}
	}
	return false
}

func (auth *AuthInfo) ValidateResources(resources []string) ([]string, bool) {
	if !auth.ResourceEnabled {
		return resources, false
	}

	var missing []string
	for _, resource := range resources {
		if !auth.ValidateResource(resource) {
			missing = append(missing, resource)
		}
	}
	return missing, len(missing) == 0
}

func (auth *AuthInfo) Merge(other *AuthInfo) error {
	if auth.Subject != other.Subject {
		return errors.New("subject mismatch")
	}

	if other.ScopeEnabled {
		auth.minimizeScopes(other)
	}
	if other.ResourceEnabled {
		auth.minimizeResources(other)
	}
	return nil
}

func (auth *AuthInfo) minimizeScopes(other *AuthInfo) {
	missingScope, ok := other.ValidateScopes(auth.Scopes)
	if !ok {
		// Remove missing scopes from auth.Scopes
		auth.removeScopes(missingScope)
	}
}

func (auth *AuthInfo) removeScopes(scopes []string) {
	for _, scope := range scopes {
		auth.removeScope(scope)
	}
}

func (auth *AuthInfo) removeScope(scope string) {
	for i, s := range auth.Scopes {
		if s == scope {
			auth.Scopes = append(auth.Scopes[:i], auth.Scopes[i+1:]...)
			return
		}
	}
}

func (auth *AuthInfo) minimizeResources(other *AuthInfo) {
	missingResource, ok := other.ValidateResources(auth.Resources)
	if !ok {
		// Remove missing resources from auth.Resources
		auth.removeResources(missingResource)
	}
}

func (auth *AuthInfo) removeResources(resources []string) {
	for _, resource := range resources {
		auth.removeResource(resource)
	}
}

func (auth *AuthInfo) removeResource(resource string) {
	for i, r := range auth.Resources {
		if r == resource {
			auth.Resources = append(auth.Resources[:i], auth.Resources[i+1:]...)
			return
		}
	}
}
