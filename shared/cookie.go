package shared

import (
	"fmt"

	"github.com/gorilla/sessions"
)

var (
	cookieSecret []byte
	// CookieStore global instance of cookie store
	CookieStore *sessions.CookieStore
)

func init() {
	fmt.Println("Initialize coockie store...")

	cookieSecret = []byte("Please use a more sensible secret than this one")
	CookieStore = sessions.NewCookieStore(cookieSecret)
}
