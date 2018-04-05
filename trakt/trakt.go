package trakt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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

func getTokenFile() string {
	var home string
	home = os.Getenv("HOME")
	if home == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		home = u.HomeDir
	}
	return filepath.Join(home, ".go.trakt.json")
}

// Load tokens from local file
func loadLocalTokens() (LocalTokens, error) {
	var tokens LocalTokens
	content, err := ioutil.ReadFile(getTokenFile())
	if err != nil {
		return tokens, fmt.Errorf("No token file")
	}
	if err := json.Unmarshal(content, &tokens); err != nil {
		log.Fatal(err)
	}
	return tokens, nil

}

// Save tokens to local file
func saveLocalTokens(t *LocalTokens) {
	fp, err := os.OpenFile(getTokenFile(), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}
	res, _ := json.Marshal(t)
	fp.Write(res)
	fp.Close()

}

func traktPOSTJSON(path string, input interface{}, output interface{}) (*http.Response, []byte) {
	postdata, _ := json.Marshal(input)
	req, err := http.Post(TraktAPIURL+path, "application/json", bytes.NewReader(postdata))
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()
	resp, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(resp, &output)
	return req, resp
}

// GetDeviceAccessToken ...
func (c *Client) GetDeviceAccessToken() string {
	var t LocalTokens
	t, err := loadLocalTokens()
	if err == nil && t.AccessToken != "" && t.RefreshToken != "" && t.ExpiresIn != 0 {
		if t.CreatedAt+t.ExpiresIn > time.Now().Add(-time.Hour*24*10).Unix() {
			return t.AccessToken
		}
	}
	if t.RefreshToken != "" {
		req, _ := traktPOSTJSON("oauth/token", AuthReq{
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
		traktPOSTJSON("oauth/device/code", AuthReq{ClientID: c.ID}, &tmp)
		start := time.Now()
		fmt.Printf("Go to %s and enter code %s, i'll keep checking until you do...\n\n\n", tmp.VerificationURL, tmp.UserCode)
		time.Sleep(5 * time.Second)
		for {
			if time.Now().Unix() >= start.Unix()+tmp.ExpiresIn {
				log.Fatal("You did not authorize, you need to start again")
			}
			rrr := AuthReq{
				ClientID:     c.ID,
				ClientSecret: c.Secret,
				Code:         tmp.DeviceCode,
			}
			fmt.Printf("oauth/device/token. %s ...\n", rrr)
			req, _ := traktPOSTJSON("oauth/device/token", rrr, &t)
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
func (c *Client) GET(path string) []byte {
	path = strings.TrimLeft(path, "/")
	req, err := http.NewRequest("GET", TraktAPIURL+path, nil)
	if err != nil {
		log.Fatal(err)
	}
	c.AddOAuthHeaders(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return data
}
