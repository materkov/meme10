package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	pb "socialnet/proto/posts"

	"github.com/twitchtv/twirp"
)

type AuthServer struct {
	// Auth-specific configuration could go here
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewAuthServer() *AuthServer {
	return &AuthServer{
		clientID:     os.Getenv("VK_CLIENT_ID"),
		clientSecret: os.Getenv("VK_CLIENT_SECRET"),
		redirectURI:  os.Getenv("VK_REDIRECT_URI"),
	}
}

// VKAuth implements the Twirp AuthService method for authenticating with VK using OAuth2.
func (s *AuthServer) VKAuth(ctx context.Context, req *pb.VKAuthRequest) (*pb.VKAuthResponse, error) {
	// Build VK OAuth2 token exchange URL
	tokenURL := "https://oauth.vk.com/access_token"
	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("client_secret", s.clientSecret)
	params.Set("redirect_uri", s.redirectURI)
	params.Set("code", req.Code)

	resp, err := http.Get(tokenURL + "?" + params.Encode())
	if err != nil {
		return nil, twirp.InternalError("Failed to request access token: " + err.Error())
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, twirp.InternalError("Failed to read response body: " + err.Error())
	}

	// Parse VK response (success or error)
	var vkResp struct {
		AccessToken      string `json:"access_token"`
		ExpiresIn        int    `json:"expires_in"`
		UserID           int64  `json:"user_id"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.Unmarshal(bodyBytes, &vkResp); err != nil {
		return nil, twirp.InternalError("Failed to decode token response: " + err.Error())
	}

	if vkResp.Error != "" {
		errorMsg := vkResp.Error
		if vkResp.ErrorDescription != "" {
			errorMsg += ": " + vkResp.ErrorDescription
		}
		return nil, twirp.InvalidArgumentError("code", errorMsg)
	}

	if vkResp.AccessToken == "" {
		return nil, twirp.InvalidArgumentError("code", "Invalid or expired authorization code")
	}

	userID := strconv.FormatInt(vkResp.UserID, 10)
	return &pb.VKAuthResponse{
		Token:  vkResp.AccessToken,
		UserId: userID,
	}, nil
}

// GetVKAuthURL implements the Twirp AuthService method for generating the VK OAuth2 URL.
func (s *AuthServer) GetVKAuthURL(ctx context.Context, req *pb.GetVKAuthURLRequest) (*pb.GetVKAuthURLResponse, error) {
	authURL := "https://oauth.vk.com/authorize"
	params := url.Values{}
	params.Set("client_id", s.clientID)
	params.Set("redirect_uri", s.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "email")

	fullURL := authURL + "?" + params.Encode()
	return &pb.GetVKAuthURLResponse{Url: fullURL}, nil
}
