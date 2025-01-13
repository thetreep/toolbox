package translator

import "github.com/nicksnyder/go-i18n/v2/i18n"

//go:generate ./gen.sh fr ./

// for testing purposes.
const (
	HelloWorld          = "helloWorld"
	HelloYou            = "helloYou"
	SurroundedByFriends = "surroundedByFriends"
)

var (
	_ = &i18n.Message{
		ID:    HelloWorld,
		Other: "Hello world!",
	}
	_ = &i18n.Message{
		ID:    HelloYou,
		Other: "Hello {{.}}!",
	}

	_ = &i18n.Message{
		ID:    SurroundedByFriends,
		Other: "surrounded by friends",
	}
)
