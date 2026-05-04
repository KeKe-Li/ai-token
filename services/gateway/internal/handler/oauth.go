package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/KeKe-Li/ai-token/services/gateway/internal/config"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/middleware"
	"github.com/KeKe-Li/ai-token/services/gateway/internal/model"
)

type OAuthHandler struct {
	userStore *model.UserStore
	cfg       *config.Config
}

func NewOAuthHandler(userStore *model.UserStore, cfg *config.Config) *OAuthHandler {
	return &OAuthHandler{userStore: userStore, cfg: cfg}
}

func (h *OAuthHandler) GitHubLogin(c *gin.Context) {
	clientID := h.cfg.GitHubClientID
	if clientID == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "GitHub OAuth 未配置"})
		return
	}
	redirectURI := fmt.Sprintf("%s/api/auth/github/callback", h.cfg.AllowedOrigins)
	url := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email", clientID, redirectURI)
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *OAuthHandler) GitHubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusFound, "/login?error=missing_code")
		return
	}

	accessToken, err := h.exchangeGitHubToken(code)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=token_exchange_failed")
		return
	}

	ghUser, err := h.getGitHubUser(accessToken)
	if err != nil {
		c.Redirect(http.StatusFound, "/login?error=user_fetch_failed")
		return
	}

	user, err := h.userStore.GetByEmail(c.Request.Context(), ghUser.Email)
	if err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte("oauth_"+ghUser.Login), bcrypt.DefaultCost)
		user, err = h.userStore.Create(c.Request.Context(), ghUser.Login, ghUser.Email, string(hash))
		if err != nil {
			c.Redirect(http.StatusFound, "/login?error=create_user_failed")
			return
		}
	}

	token, _ := middleware.GenerateToken(user.ID, user.Role, h.cfg.JWTSecret)
	c.Redirect(http.StatusFound, fmt.Sprintf("/login?token=%s&user=%s", token, ghUser.Login))
}

type githubTokenResp struct {
	AccessToken string `json:"access_token"`
}

func (h *OAuthHandler) exchangeGitHubToken(code string) (string, error) {
	req, _ := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", nil)
	q := req.URL.Query()
	q.Set("client_id", h.cfg.GitHubClientID)
	q.Set("client_secret", h.cfg.GitHubClientSecret)
	q.Set("code", code)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result githubTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access token")
	}
	return result.AccessToken, nil
}

type githubUser struct {
	Login string `json:"login"`
	Email string `json:"email"`
}

func (h *OAuthHandler) getGitHubUser(token string) (*githubUser, error) {
	req, _ := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var user githubUser
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	if user.Email == "" {
		user.Email = user.Login + "@github.users.noreply"
	}
	return &user, nil
}
