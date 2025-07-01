package clerkauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ClerkTokenResponse represents the response from Clerk token verification
type ClerkTokenResponse struct {
	Object           string   `json:"object"`
	ID               string   `json:"id"`
	ClientID         string   `json:"client_id"`
	Subject          string   `json:"subject"`
	Scopes           []string `json:"scopes"`
	Revoked          bool     `json:"revoked"`
	RevocationReason string   `json:"revocation_reason,omitempty"`
	Expired          bool     `json:"expired"`
	Expiration       int64    `json:"expiration"`
	CreatedAt        int64    `json:"created_at"`
	UpdatedAt        int64    `json:"updated_at"`
}

// SessionTokenClaims represents claims from Clerk JWT session token
type SessionTokenClaims struct {
	Sub string `json:"sub"` // User ID
	Iss string `json:"iss"` // Issuer
	Aud string `json:"aud"` // Audience
	Azp string `json:"azp"` // Authorized party
	Sid string `json:"sid"` // Session ID
	Exp int64  `json:"exp"` // Expiration time
	Iat int64  `json:"iat"` // Issued at
	Nbf int64  `json:"nbf"` // Not before
	jwt.RegisteredClaims
}

// ClerkTokenRequest represents the request payload for token verification
type ClerkTokenRequest struct {
	AccessToken string `json:"access_token"`
}

// ClerkErrorResponse represents error response from Clerk API
type ClerkErrorResponse struct {
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	ClerkTraceID string `json:"clerk_trace_id,omitempty"`
}

// JWKS represents JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"` // Key Type
	Use string `json:"use"` // Public Key Use
	Kid string `json:"kid"` // Key ID
	N   string `json:"n"`   // Modulus
	E   string `json:"e"`   // Exponent
	Alg string `json:"alg"` // Algorithm
}

// ClerkUser represents user data from Clerk API
type ClerkUser struct {
	ID             string `json:"id"`
	Object         string `json:"object"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ImageURL       string `json:"image_url"`
	HasImage       bool   `json:"has_image"`
	EmailAddresses []struct {
		ID           string `json:"id"`
		Object       string `json:"object"`
		EmailAddress string `json:"email_address"`
		Reserved     bool   `json:"reserved"`
		Verification struct {
			Status   string `json:"status"`
			Strategy string `json:"strategy"`
		} `json:"verification"`
	} `json:"email_addresses"`
}

type ClerkService interface {
	VerifyClerkToken(token string) (*jwt.Token, error)
	GetUser(userID string) (*ClerkUser, error)
}

type clerkService struct {
	secretKey   string
	frontendAPI string
	httpClient  *http.Client
	debug       bool
}

func NewService() ClerkService {
	frontendAPI := os.Getenv("CLERK_FRONTEND_API_URL")

	return &clerkService{
		secretKey:   os.Getenv("CLERK_SECRET_KEY"),
		frontendAPI: frontendAPI,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		debug: os.Getenv("DEBUG") == "true",
	}
}

var JWKs *keyfunc.JWKS

func InitClerkJWKs() error {
	frontendAPI := os.Getenv("CLERK_FRONTEND_API_URL")
	if frontendAPI == "" {
		return errors.New("CLERK_FRONTEND_API_URL environment variable not set")
	}

	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", frontendAPI)

	var err error
	JWKs, err = keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshTimeout:  time.Minute,
	})
	return err
}

func (s *clerkService) VerifyClerkToken(tokenString string) (*jwt.Token, error) {
	fmt.Println("tokenString", tokenString)
	if s.debug {
		fmt.Printf("VerifyClerkToken called, JWKs status: %v\n", JWKs != nil)
		fmt.Printf("Frontend API: %s\n", s.frontendAPI)
	}

	if JWKs == nil {
		if s.debug {
			fmt.Println("JWKs is nil, attempting to reinitialize...")
		}
		// 嘗試重新初始化
		if err := InitClerkJWKs(); err != nil {
			if s.debug {
				fmt.Printf("Failed to reinitialize JWKs: %v\n", err)
			}
			return nil, fmt.Errorf("JWKs not initialized and failed to initialize: %w", err)
		}
		if s.debug {
			fmt.Println("JWKs reinitialized successfully")
		}
	}

	token, err := jwt.Parse(tokenString, JWKs.Keyfunc)
	if err != nil {
		if s.debug {
			fmt.Printf("Failed to parse token: %v\n", err)
		}
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		if s.debug {
			fmt.Println("Token is not valid")
		}
		return nil, errors.New("invalid token")
	}

	// 驗證 claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if s.debug {
			fmt.Printf("Token claims: %+v\n", claims)
		}

		// 檢查 issuer (從 JWT token 中取得實際的 issuer)
		if iss, ok := claims["iss"].(string); ok {
			if s.debug {
				fmt.Printf("Token issuer: %s\n", iss)
			}
			// 如果有設定 frontendAPI，則驗證 issuer
			if s.frontendAPI != "" {
				expectedIssuer := fmt.Sprintf("https://%s", s.frontendAPI)
				if iss != expectedIssuer {
					if s.debug {
						fmt.Printf("Issuer mismatch: expected %s, got %s\n", expectedIssuer, iss)
					}
					return nil, fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, iss)
				}
			}
		}

		// 檢查過期時間
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				if s.debug {
					fmt.Println("Token expired")
				}
				return nil, errors.New("token expired")
			}
		}

		// 檢查 not before 時間
		if nbf, ok := claims["nbf"].(float64); ok {
			if time.Now().Unix() < int64(nbf) {
				if s.debug {
					fmt.Println("Token not yet valid")
				}
				return nil, errors.New("token not yet valid")
			}
		}
	}

	return token, nil
}

func (s *clerkService) GetUser(userID string) (*ClerkUser, error) {
	if s.secretKey == "" {
		return nil, errors.New("CLERK_SECRET_KEY not set")
	}

	url := fmt.Sprintf("https://api.clerk.com/v1/users/%s", userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("clerk API returned status %d", resp.StatusCode)
	}

	var user ClerkUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}
