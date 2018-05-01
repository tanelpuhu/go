# pushover

	import "github.com/tanelpuhu/go/pushover"
	...
	po := pushover.New(UserKey, APIToken)
	po.Send("Title...", "This is message....")
	po.SendWithURL("Title 2", "Click the url", "https://example.com/")
	...
