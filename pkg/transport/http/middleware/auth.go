package middleware

import (
	"errors"
	"net/http"

	"github.com/ElfAstAhe/go-service-template/pkg/auth"
	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	pkghlpjwt "github.com/ElfAstAhe/go-service-template/pkg/helper/jwt"
	"github.com/ElfAstAhe/go-service-template/pkg/utils"
)

// ErrCredentialsNotFound сигнализирует конвейеру, что в запросе отсутствуют
// учётные данные для данного способа аутентификации (нужно попробовать следующий).
var ErrCredentialsNotFound = errors.New("credentials not found in request")

// HTTPCredentialsExtractor извлекает сырые учётные данные из HTTP-запроса.
type HTTPCredentialsExtractor func(r *http.Request) (any, error)

// HTTPAuthProvider связывает извлечение кредов из HTTP-запроса и их проверку через auth.Authenticator.
type HTTPAuthProvider struct {
	Extract       HTTPCredentialsExtractor
	Authenticator auth.Authenticator
}

// HTTPAuthErrorHandler позволяет кастомизировать формат JSON/текстового ответа при ошибке 401.
type HTTPAuthErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// AuthMiddleware создаёт стандартный net/http middleware для аутентификации запросов.
// Если strict == false, запросы без учётных данных пропускаются дальше как auth.Guest.
func AuthMiddleware(
	providers []HTTPAuthProvider,
	strict bool,
	errHandler HTTPAuthErrorHandler,
) func(http.Handler) http.Handler {
	if errHandler == nil {
		errHandler = defaultHTTPAuthErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если субъект уже установлен выше по цепочке — просто пропускаем дальше
			subj, err := auth.FromContext(r.Context())
			if err == nil && !utils.IsNil(subj) && subj.IsAuthenticated() {
				next.ServeHTTP(w, r)
				return
			}

			for _, p := range providers {
				if p.Extract == nil || p.Authenticator == nil {
					continue
				}

				creds, err := p.Extract(r)
				if err != nil {
					if errors.Is(err, ErrCredentialsNotFound) {
						continue // Кредов этого типа нет — пробуем следующий провайдер
					}
					errHandler(w, r, err)
					return
				}

				subject, err := p.Authenticator.Authenticate(r.Context(), creds)
				if err != nil {
					errHandler(w, r, err)
					return
				}

				subjCtx := auth.WithSubject(r.Context(), subject)
				next.ServeHTTP(w, r.WithContext(subjCtx))
				return
			}

			// Ни один провайдер не нашёл учётных данных в запросе
			if strict {
				errHandler(w, r, errs.NewUtlAuthError("authentication credentials required", nil))
				return
			}

			guestCtx := auth.WithSubject(r.Context(), auth.Guest)
			next.ServeHTTP(w, r.WithContext(guestCtx))
		})
	}
}

func defaultHTTPAuthErrorHandler(w http.ResponseWriter, _ *http.Request, _ error) {
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

// ====================================================================
// Готовые экстракторы под JWT, Basic и Cert (mTLS)
// ====================================================================

// NewHTTPJWTExtractor адаптирует pkghlpjwt.HTTPExtractor под конвейер AuthMiddleware.
func NewHTTPJWTExtractor(extractor *pkghlpjwt.HTTPExtractor, headerName, cookieName string) HTTPCredentialsExtractor {
	return func(r *http.Request) (any, error) {
		token, err := extractor.TokenFromRequestCookie(r, cookieName)
		if err != nil {
			token, err = extractor.TokenFromRequestHeader(r, headerName)
			if err != nil {
				return nil, ErrCredentialsNotFound
			}
		}

		return &auth.JWTCredentials{
			Token: token,
		}, nil
	}
}

// NewHTTPBasicExtractor извлекает пару логин/пароль из заголовка Authorization: Basic.
func NewHTTPBasicExtractor() HTTPCredentialsExtractor {
	return func(r *http.Request) (any, error) {
		username, password, ok := r.BasicAuth()
		if !ok {
			return nil, ErrCredentialsNotFound
		}

		return &auth.BasicCredentials{
			Username: username,
			Password: password,
		}, nil
	}
}

// NewHTTPCertExtractor извлекает проверенный клиентский TLS-сертификат из соединения.
func NewHTTPCertExtractor() HTTPCredentialsExtractor {
	return func(r *http.Request) (any, error) {
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			return nil, ErrCredentialsNotFound
		}

		return &auth.CertCredentials{
			Certificate: r.TLS.PeerCertificates[0],
		}, nil
	}
}
