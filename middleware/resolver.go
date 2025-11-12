package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// TenantContextKey is the key used to store tenant information in fiber context
const TenantContextKey = "tenant"

// TenantResolver is a function that extracts tenant identifier from the request
type TenantResolver func(c *fiber.Ctx) (string, error)

// SubdomainResolver extracts tenant from subdomain (e.g., tenant1.example.com -> tenant1)
func SubdomainResolver(c *fiber.Ctx) (string, error) {
	host := c.Hostname()

	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	parts := strings.Split(host, ".")

	// Need at least 2 parts for a subdomain
	// For localhost: tenant.localhost (2 parts is ok)
	// For domains: tenant.example.com (3+ parts required)
	if len(parts) >= 2 {
		subdomain := parts[0]

		// Filter out common non-tenant subdomains
		if subdomain == "www" || subdomain == "api" || subdomain == "localhost" {
			return "", fiber.NewError(fiber.StatusBadRequest, "No valid tenant subdomain found")
		}

		// For 2-part hosts, only accept if second part is "localhost"
		if len(parts) == 2 && parts[1] != "localhost" {
			return "", fiber.NewError(fiber.StatusBadRequest, "No valid tenant subdomain found")
		}

		// Valid subdomain found
		return subdomain, nil
	}

	return "", fiber.NewError(fiber.StatusBadRequest, "No valid tenant subdomain found")
}

// HeaderResolver extracts tenant from a custom header
func HeaderResolver(headerName string) TenantResolver {
	return func(c *fiber.Ctx) (string, error) {
		tenant := c.Get(headerName)
		if tenant == "" {
			return "", fiber.NewError(fiber.StatusBadRequest, "Tenant header not found")
		}
		return tenant, nil
	}
}

// PathPrefixResolver extracts tenant from URL path prefix (e.g., /tenant1/users -> tenant1)
func PathPrefixResolver(c *fiber.Ctx) (string, error) {
	path := c.Path()
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	if len(parts) > 0 && parts[0] != "" {
		return parts[0], nil
	}

	return "", fiber.NewError(fiber.StatusBadRequest, "No tenant found in path")
}

// QueryParamResolver extracts tenant from query parameter
func QueryParamResolver(paramName string) TenantResolver {
	return func(c *fiber.Ctx) (string, error) {
		tenant := c.Query(paramName)
		if tenant == "" {
			return "", fiber.NewError(fiber.StatusBadRequest, "Tenant query parameter not found")
		}
		return tenant, nil
	}
}

// ChainResolvers tries multiple resolvers in order until one succeeds
func ChainResolvers(resolvers ...TenantResolver) TenantResolver {
	return func(c *fiber.Ctx) (string, error) {
		for _, resolver := range resolvers {
			tenant, err := resolver(c)
			if err == nil && tenant != "" {
				return tenant, nil
			}
		}
		return "", fiber.NewError(fiber.StatusBadRequest, "No tenant found using any resolver")
	}
}

// CustomResolver allows you to provide your own resolution logic
func CustomResolver(fn func(c *fiber.Ctx) (string, error)) TenantResolver {
	return fn
}
