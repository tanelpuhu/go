package trakt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tanelpuhu/go/paths"
)

var (
	// TraktAPIURL ...
	TraktAPIURL = "https://api.trakt.tv/"
)

// AuthReq ...
type AuthReq struct {
	ClientID        string `json:"client_id,omitempty"`
	ClientSecret    string `json:"client_secret,omitempty"`
	Code            string `json:"code,omitempty"`
	DeviceCode      string `json:"device_code,omitempty"`
	UserCode        string `json:"user_code,omitempty"`
	VerificationURL string `json:"verification_url,omitempty"`
	ExpiresIn       int64  `json:"expires_in,omitempty"`
	Interval        int    `json:"interval,omitempty"`
	RefreshToken    string `json:"refresh_token,omitempty"`
	RedirectURI     string `json:"redirect_uri,omitempty"`
	GrantType       string `json:"grant_type,omitempty"`
}

// LocalTokens ...
type LocalTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	CreatedAt    int64  `json:"created_at"`
}

// Client ...
type Client struct {
	ID     string
	Secret string
	Tokens LocalTokens
}

// New ...
func New(id, secret string) *Client {
	return &Client{
		ID:     id,
		Secret: secret,
	}
}

func getLocalTokensFileName() string {
	home, err := paths.GetHome()
	if err != nil {
		log.Fatalf("error getting home path: %v", err)
	}
	return filepath.Join(home, ".go.trakt.json")
}

// Load tokens from local file
func loadLocalTokens() (LocalTokens, error) {
	var tokens LocalTokens
	content, err := ioutil.ReadFile(getLocalTokensFileName())
	if err != nil {
		return tokens, fmt.Errorf("no token file")
	}
	if err := json.Unmarshal(content, &tokens); err != nil {
		log.Fatalf("error decoding local token file: %v", err)
	}
	return tokens, nil

}

// Save tokens to local file
func saveLocalTokens(t *LocalTokens) {
	fp, err := os.OpenFile(getLocalTokensFileName(), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("error opening local token file: %v", err)
	}
	res, err := json.Marshal(t)
	if err != nil {
		log.Fatalf("error encoding token for local file: %v", err)
	}
	fp.Write(res)
	fp.Close()

}

func traktAuthPOST(path string, input interface{}, output interface{}) (*http.Response, []byte) {
	postdata, _ := json.Marshal(input)
	req, err := http.Post(TraktAPIURL+path, "application/json", bytes.NewReader(postdata))
	if err != nil {
		log.Fatalf("error making request: %v", err)
	}
	defer req.Body.Close()
	resp, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatalf("error reading response: %v", err)
	}
	json.Unmarshal(resp, &output)
	return req, resp
}

// GetDeviceAccessToken ...
func (c *Client) GetDeviceAccessToken() string {
	var t LocalTokens
	t, err := loadLocalTokens()
	if err == nil && t.AccessToken != "" && t.RefreshToken != "" && t.ExpiresIn != 0 {
		if t.CreatedAt+t.ExpiresIn > time.Now().Add(time.Hour*24*7).Unix() {
			return t.AccessToken
		}
	}
	if t.RefreshToken != "" {
		req, _ := traktAuthPOST("oauth/token", AuthReq{
			ClientID:     c.ID,
			ClientSecret: c.Secret,
			RefreshToken: t.RefreshToken,
			RedirectURI:  "urn:ietf:wg:oauth:2.0:oob",
			GrantType:    "refresh_token",
		}, &t)
		if req.StatusCode == 200 {
			saveLocalTokens(&t)
		}
	} else {
		var tmp AuthReq
		traktAuthPOST("oauth/device/code", AuthReq{ClientID: c.ID}, &tmp)
		start := time.Now()
		fmt.Printf("Go to %s and enter code %s, i'll keep checking until you do...\n\n\n", tmp.VerificationURL, tmp.UserCode)
		time.Sleep(5 * time.Second)
		for {
			if time.Now().Unix() >= start.Unix()+tmp.ExpiresIn {
				log.Fatal("You did not authorize, you need to start again")
			}
			req, _ := traktAuthPOST("oauth/device/token", AuthReq{
				ClientID:     c.ID,
				ClientSecret: c.Secret,
				Code:         tmp.DeviceCode,
			}, &t)
			if req.StatusCode == 200 {
				saveLocalTokens(&t)
				break
			} else if req.StatusCode == 400 {
				fmt.Println("Pending - waiting for the user to authorize your app")
			} else if req.StatusCode == 404 {
				fmt.Println("Not Found - invalid device_code")
				break
			} else if req.StatusCode == 409 {
				fmt.Println("Already Used - user already approved this code")
				break
			} else if req.StatusCode == 410 {
				fmt.Println("Expired - the tokens have expired, restart the process")
				break
			} else if req.StatusCode == 418 {
				fmt.Println("Denied - user explicitly denied this code")
				break
			} else if req.StatusCode == 429 {
				fmt.Println("Slow Down - your app is polling too quickly")
				time.Sleep(5 * time.Second)
			}
			time.Sleep(5 * time.Second)
		}
	}
	return t.AccessToken
}

// AddOAuthHeaders ...
func (c *Client) AddOAuthHeaders(req *http.Request) {
	accessToken := c.GetDeviceAccessToken()
	if accessToken == "" {
		log.Fatal("No AccessToken found")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("trakt-api-key", c.ID)
	req.Header.Add("trakt-api-version", "2")
	req.Header.Add("Authorization", "Bearer "+accessToken)
}

// GET ...
func (c *Client) GET(path string, output interface{}) error {
	return makeRequest(c, http.MethodGet, path, nil, &output)
}

// POST ...
func (c *Client) POST(path string, input, output interface{}) error {
	return makeRequest(c, http.MethodPost, path, &input, &output)
}

func makeRequest(c *Client, method, path string, input, output *interface{}) error {
	path = strings.TrimLeft(path, "/")
	var body io.Reader
	if input != nil {
		postdata, err := json.Marshal(input)
		if err != nil {
			log.Fatalf("error encoding to json: %v", err)
		}
		body = bytes.NewReader(postdata)
	}
	req, err := http.NewRequest(method, TraktAPIURL+path, body)
	if err != nil {
		log.Fatalf("error creating request: %v", err)
	}
	c.AddOAuthHeaders(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("error making request: %v", err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error reading response: %v", err)
	}
	if err := json.Unmarshal(content, &output); err != nil {
		log.Fatalf("error decoding response: %v - %s", err, content)
	}
	return nil
}
