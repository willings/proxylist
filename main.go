package proxylist

import (
	"encoding/json"
	"fmt"
	"github.com/willings/proxypool/provider"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"strings"
)

const AUTH_ON = false

const API_KEY = "da39a3ee5e6b4b0d3255bfef95601890afd80709"


func init() {
	http.HandleFunc("/", handle)
	http.HandleFunc("/proxylist/refresh.json", handle_refresh_json)
}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1><It Works!/h1>")
}

func handle_refresh_json(w http.ResponseWriter, r *http.Request) {
	if AUTH_ON {
		apikey := r.URL.Query().Get("apikey") // TODO
		if apikey != API_KEY {
			errJson(w, ACCESS_DENIED)
			return
		}
	}

	providers := r.URL.Query().Get("providers")
	providerArr := strings.Split(providers, "|")

	loaders := make([]provider.ProxyProvider, 0)

	for _, name := range providerArr {
		if name == "any" {
			loaders = provider.CreateAllProvider()
			break
		} else {
			loader := provider.CreateProvider(name)
			if loader != nil {
				loaders = append(loaders, loader)
			}
		}

	}

	if len(loaders) == 0 {
		errJson(w, NO_PROVIDER)
		return
	}

	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)

	p := provider.CreateMultiLoader(loaders...)
	p.SetClient(client)
	items, _ := p.Load()

	ret, _ := json.Marshal(items)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(ret))
}
