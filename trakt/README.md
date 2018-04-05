# trakt

	...
	trkt := trakt.New(ClientID, ClientSecret)
	req, err := http.NewRequest("GET", "https://api.trakt.tv/users/me/watchlist/movies", nil)
	if err != nil {
		logrus.Fatal(err)
	}
	trkt.AddOAuthHeaders(req)
	...