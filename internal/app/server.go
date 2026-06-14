package app

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/ward-sentry/chur-keenetic/internal/amneziawg"
	"github.com/ward-sentry/chur-keenetic/internal/buildinfo"
	"github.com/ward-sentry/chur-keenetic/internal/system"
	"github.com/ward-sentry/chur-keenetic/internal/web"
)

func NewServer(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "index.html", time.Time{}, web.Index())
	})
	mux.Handle("GET /icons/", http.StripPrefix("/icons/", http.FileServer(http.FS(web.Icons()))))

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"version": buildinfo.Version,
			"commit":  buildinfo.Commit,
		})
	})

	mux.HandleFunc("GET /api/system", func(w http.ResponseWriter, r *http.Request) {
		report := system.Collect(r.Context())
		writeJSON(w, http.StatusOK, report)
	})

	mux.HandleFunc("POST /api/runtime/amneziawg/install", func(w http.ResponseWriter, r *http.Request) {
		result := system.InstallAmneziaWG(r.Context())
		status := http.StatusOK
		if !result.Ready {
			status = http.StatusInternalServerError
		}
		writeJSON(w, status, result)
	})

	mux.HandleFunc("GET /api/providers/amneziawg/configs", func(w http.ResponseWriter, r *http.Request) {
		configs, err := amneziawg.ListConfigs(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"configs": configs,
		})
	})

	mux.HandleFunc("POST /api/providers/amneziawg/configs", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(512 * 1024); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		name := r.FormValue("name")
		file, _, err := r.FormFile("config")
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		defer file.Close()

		content, err := io.ReadAll(io.LimitReader(file, 256*1024+1))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		config, err := amneziawg.SaveConfig(r.Context(), amneziawg.SaveConfigRequest{
			Name:        name,
			Description: r.FormValue("description"),
			Content:     string(content),
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		writeJSON(w, http.StatusCreated, config)
	})

	mux.HandleFunc("PUT /api/providers/amneziawg/configs/{name}", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(512 * 1024); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		var content string
		file, _, err := r.FormFile("config")
		if err == nil {
			defer file.Close()
			fileContent, err := io.ReadAll(io.LimitReader(file, 256*1024+1))
			if err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			content = string(fileContent)
		}

		result, err := amneziawg.UpdateConfig(r.Context(), amneziawg.UpdateConfigRequest{
			Name:        r.PathValue("name"),
			Description: r.FormValue("description"),
			MTU:         r.FormValue("mtu"),
			Content:     content,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error":  err.Error(),
				"result": result,
			})
			return
		}

		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("GET /api/providers/amneziawg/configs/{name}/status", func(w http.ResponseWriter, r *http.Request) {
		status, err := amneziawg.Status(r.Context(), r.PathValue("name"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, status)
	})

	mux.HandleFunc("POST /api/providers/amneziawg/configs/{name}/start", func(w http.ResponseWriter, r *http.Request) {
		result, err := amneziawg.Start(r.Context(), r.PathValue("name"))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error":  err.Error(),
				"result": result,
			})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("POST /api/providers/amneziawg/configs/{name}/stop", func(w http.ResponseWriter, r *http.Request) {
		result, err := amneziawg.Stop(r.Context(), r.PathValue("name"))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error":  err.Error(),
				"result": result,
			})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	mux.HandleFunc("DELETE /api/providers/amneziawg/configs/{name}", func(w http.ResponseWriter, r *http.Request) {
		result, err := amneziawg.DeleteConfig(r.Context(), r.PathValue("name"))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"error":  err.Error(),
				"result": result,
			})
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	return loggingMiddleware(logger, mux)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"duration", time.Since(started).String(),
		)
	})
}
