package pushover

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// APIUrl ...
const APIUrl = "https://api.pushover.net/1/messages.json"

// Client ...
type Client struct {
	User   string
	Token  string
	Device string
	Sound  string
}

// Message ...
type Message struct {
	User     string `json:"user"`
	Token    string `json:"token"`
	Device   string `json:"device,omitempty"`
	Sound    string `json:"sound,omitempty"`
	Title    string `json:"title,omitempty"`
	Message  string `json:"message"`
	URL      string `json:"url,omitempty"`
	URLTitle string `json:"url_title,omitempty"`
}

// Response ...
type Response struct {
	Status  int      `json:"status,omitempty"`
	Request string   `json:"request,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

// New ...
func New(user, token string) *Client {
	return &Client{
		User:   user,
		Token:  token,
		Device: "",
	}
}

// SetDevice ...
func (c *Client) SetDevice(device string) {
	c.Device = device
}

// SetSound ...
func (c *Client) SetSound(sound string) {
	c.Sound = sound
}

// Send ...
func (c *Client) Send(title, message string) (Response, error) {
	return send(c, title, message, "", "")
}

// SendWithURL ...
func (c *Client) SendWithURL(title, message, url string) (Response, error) {
	return send(c, title, message, url, "")
}

// SendWithURLTitle ...
func (c *Client) SendWithURLTitle(title, message, url, urltitle string) (Response, error) {
	return send(c, title, message, url, urltitle)
}

func send(c *Client, title, message, url, urltitle string) (Response, error) {
	response := Response{}
	params, err := json.Marshal(Message{
		User:     c.User,
		Token:    c.Token,
		Device:   c.Device,
		Title:    title,
		Message:  message,
		URL:      url,
		URLTitle: urltitle,
	})
	if err != nil {
		return response, err
	}
	req, err := http.Post(APIUrl, "application/json", bytes.NewBuffer(params))
	if err != nil {
		return response, err
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return response, err
	}
	defer req.Body.Close()
	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, err
	}
	if req.StatusCode != http.StatusOK {
		return response, fmt.Errorf("Response: %s", body)
	}
	return response, nil
}
