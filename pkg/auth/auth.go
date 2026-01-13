package auth

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"
)

// getTokenPrefix возвращает префикс токена для безопасного логирования
func getTokenPrefix(token string, length int) string {
	if len(token) == 0 {
		return ""
	}
	if len(token) <= length {
		return token
	}
	return token[:length] + "..."
}

// BearerAuthMiddleware создает middleware для проверки Bearer токена
// Если token пустой, проверка пропускается (для разработки/тестирования)
func BearerAuthMiddleware(logger *slog.Logger, token string) func(http.HandlerFunc) http.HandlerFunc {
	// Логируем состояние токена при создании middleware
	if token == "" {
		logger.Warn("BearerAuthMiddleware created with empty token - authorization DISABLED",
			"event", "auth_middleware_created_disabled",
		)
	} else {
		logger.Info("BearerAuthMiddleware created with token - authorization ENABLED",
			"event", "auth_middleware_created_enabled",
			"token_length", len(token),
			"token_prefix", getTokenPrefix(token, 20),
		)
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Если токен не настроен, пропускаем проверку (для разработки/тестирования)
			if token == "" {
				logger.Warn("Bearer token not configured, skipping authorization check",
					"event", "auth_disabled",
					"path", r.URL.Path,
				)
				next(w, r)
				return
			}

			// Получаем токен из заголовка Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Missing Authorization header",
					"event", "auth_missing",
					"path", r.URL.Path,
				)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Проверяем формат Bearer токена
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				logger.Warn("Invalid Authorization header format",
					"event", "auth_invalid_format",
					"path", r.URL.Path,
					"header_prefix", getTokenPrefix(authHeader, 20),
				)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Извлекаем токен
			providedToken := strings.TrimPrefix(authHeader, bearerPrefix)

			// Проверяем токен
			// Используем constant-time comparison для защиты от timing attacks
			// subtle.ConstantTimeCompare возвращает 1 если строки равны, 0 если нет
			if subtle.ConstantTimeCompare([]byte(providedToken), []byte(token)) != 1 {
				logger.Warn("Invalid Bearer token",
					"event", "auth_invalid_token",
					"token_length", len(providedToken),
					"expected_length", len(token),
					"provided_token_prefix", getTokenPrefix(providedToken, 20),
					"expected_token_prefix", getTokenPrefix(token, 20),
				)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Логируем успешную авторизацию для отладки
			logger.Debug("Bearer token validated successfully",
				"event", "auth_success",
				"token_length", len(providedToken),
			)

			// Токен валиден, продолжаем обработку
			next(w, r)
		}
	}
}
