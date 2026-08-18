package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"windops/internal/application"
	"windops/internal/fault"
)

type Server struct {
	coordinator *application.Coordinator
	webRoot     string
	logger      *slog.Logger
	mux         *http.ServeMux
}

func New(coordinator *application.Coordinator, webRoot string, logger *slog.Logger) *Server {
	s := &Server{coordinator: coordinator, webRoot: webRoot, logger: logger, mux: http.NewServeMux()}
	s.routes()
	return s
}
func (s *Server) Handler() http.Handler { return s.recover(s.requestID(s.log(s.mux))) }
func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.health)
	s.mux.HandleFunc("GET /readyz", s.ready)
	s.mux.HandleFunc("GET /api/rules", s.rules)
	s.mux.HandleFunc("GET /api/overview", s.overview)
	s.mux.HandleFunc("POST /api/decisions", s.decision)
	s.mux.HandleFunc("GET /api/farms", s.entities("farms"))
	s.mux.HandleFunc("GET /api/campaigns", s.entities("campaigns"))
	s.mux.HandleFunc("GET /api/permits", s.entities("permits"))
	s.mux.HandleFunc("GET /api/work-orders", s.entities("work_orders"))
	s.mux.HandleFunc("GET /api/dispatches", s.entities("dispatches"))
	s.mux.Handle("/", http.HandlerFunc(s.frontend))
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"status": "ok", "time": time.Now().UTC()})
}
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if err := s.coordinator.DB.Ping(r.Context()); err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ready"})
}
func (s *Server) rules(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"items": application.Rules()})
}
func (s *Server) overview(w http.ResponseWriter, r *http.Request) {
	tenant := tenantFrom(r)
	result, err := s.coordinator.Overview(r.Context(), tenant)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, 200, result)
}
func (s *Server) decision(w http.ResponseWriter, r *http.Request) {
	var request application.DecisionRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	if request.Context.TenantID == "" {
		request.Context.TenantID = tenantFrom(r)
	}
	if request.RequestID == "" {
		request.RequestID = requestIDFrom(r)
	}
	result, err := s.coordinator.Evaluate(r.Context(), request)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, 200, result)
}
func (s *Server) entities(table string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowed := map[string]bool{"farms": true, "campaigns": true, "permits": true, "work_orders": true, "dispatches": true}
		if !allowed[table] {
			writeError(w, r, fault.New(fault.CodeNotFound, "http.entities", "resource was not found"))
			return
		}
		rows, err := s.coordinator.DB.SQL().QueryContext(r.Context(), "SELECT data_json FROM "+table+" WHERE tenant_id=? ORDER BY updated_at DESC,id ASC LIMIT 200", tenantFrom(r))
		if err != nil {
			writeError(w, r, err)
			return
		}
		defer rows.Close()
		items := []json.RawMessage{}
		for rows.Next() {
			var raw []byte
			if err := rows.Scan(&raw); err != nil {
				writeError(w, r, err)
				return
			}
			items = append(items, json.RawMessage(raw))
		}
		if err := rows.Err(); err != nil {
			writeError(w, r, err)
			return
		}
		writeJSON(w, 200, map[string]any{"items": items, "total": len(items)})
	}
}
func (s *Server) frontend(w http.ResponseWriter, r *http.Request) {
	clean := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if clean == "." {
		clean = "index.html"
	}
	target := filepath.Join(s.webRoot, clean)
	if info, err := os.Stat(target); err == nil && !info.IsDir() {
		http.ServeFile(w, r, target)
		return
	}
	index := filepath.Join(s.webRoot, "index.html")
	if _, err := os.Stat(index); err == nil {
		http.ServeFile(w, r, index)
		return
	}
	http.NotFound(w, r)
}
func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fault.Wrap(fault.CodeInvalid, "http.decode", "invalid JSON body", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fault.New(fault.CodeInvalid, "http.decode", "request body must contain one JSON value")
	}
	return nil
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	code := fault.CodeOf(err)
	status := 500
	switch code {
	case fault.CodeInvalid:
		status = 400
	case fault.CodeNotFound:
		status = 404
	case fault.CodeConflict, fault.CodePrecondition, fault.CodeCapacity:
		status = 409
	case fault.CodeUnavailable:
		status = 503
	case fault.CodeCanceled:
		status = 499
	}
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": string(code), "message": err.Error(), "request_id": requestIDFrom(r)}})
}

type contextKey string

const requestIDKey contextKey = "request-id"

func requestIDFrom(r *http.Request) string {
	if value, ok := r.Context().Value(requestIDKey).(string); ok {
		return value
	}
	return "unknown"
}
func tenantFrom(r *http.Request) string {
	if value := strings.TrimSpace(r.Header.Get("X-Tenant-ID")); value != "" {
		return value
	}
	return "demo"
}
func (s *Server) requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if id == "" {
			id = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}
func (s *Server) log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("request", "method", r.Method, "path", r.URL.Path, "request_id", requestIDFrom(r), "elapsed", time.Since(started))
	})
}
func (s *Server) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if value := recover(); value != nil {
				s.logger.Error("panic recovered", "value", value, "request_id", requestIDFrom(r))
				writeError(w, r, fault.New(fault.CodeInternal, "http.recover", "internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
