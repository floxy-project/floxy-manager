package middlewares

import (
	"net/http"
	"strings"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/internal/repository/auditlog"
	"github.com/rom8726/floxy-manager/pkg/db"
)

func AuditMiddleware(executor db.Tx) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isMutatingMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			username := appcontext.Username(ctx)
			if username == "" {
				next.ServeHTTP(w, r)
				return
			}

			entity, entityID := extractEntityFromPath(r.URL.Path)
			if entity == "" || entityID == "" {
				next.ServeHTTP(w, r)
				return
			}

			action := mapMethodToAction(r.Method)

			_ = auditlog.WriteLog(ctx, executor, entity, entityID, action)

			next.ServeHTTP(w, r)
		})
	}
}

func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func mapMethodToAction(method string) string {
	switch method {
	case http.MethodPost:
		return domain.ActionCreate
	case http.MethodPut, http.MethodPatch:
		return domain.ActionUpdate
	case http.MethodDelete:
		return domain.ActionDelete
	default:
		return "unknown"
	}
}

func extractEntityFromPath(path string) (entity, entityID string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "", ""
	}

	if parts[0] == "api" {
		if len(parts) >= 2 && parts[1] == "v1" {
			if len(parts) >= 4 {
				entity = parts[2]
				entityID = parts[3]
				if entity == "projects" && len(parts) >= 5 {
					if parts[4] == "memberships" && len(parts) >= 6 {
						entity = domain.EntityMembership
						entityID = parts[5]
					} else if parts[4] == "workflows" {
						entity = domain.EntityWorkflow
					}
				}
			} else if len(parts) >= 3 {
				entity = parts[2]
			}
		} else {
			if len(parts) >= 3 {
				entity = domain.EntityPlugin
				entityID = parts[2]
			} else if len(parts) >= 2 {
				entity = domain.EntityPlugin
			}
		}
	} else {
		entity = parts[0]
		if len(parts) >= 2 {
			entityID = parts[1]
		}
	}

	if entityID == "" {
		entityID = extractIDFromPath(path)
	}

	return entity, entityID
}

func extractIDFromPath(path string) string {
	patterns := []struct {
		pattern string
		prefix  string
	}{
		{"/instances/", "instance"},
		{"/workflows/", "workflow"},
		{"/decisions/", "decision"},
		{"/dlq/", "dlq"},
	}

	for _, p := range patterns {
		if strings.Contains(path, p.pattern) {
			parts := strings.Split(path, p.pattern)
			if len(parts) > 1 {
				idParts := strings.Split(parts[1], "/")
				if len(idParts) > 0 && idParts[0] != "" {
					return idParts[0]
				}
			}
		}
	}

	return ""
}
