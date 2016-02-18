package proxylist

import (
	"encoding/json"
	"fmt"
	"github.com/willings/proxypool/provider"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"strings"
	"time"
)

const AUTH_ON = false

const API_KEY = "YOUR_SEC_KEY"

const (
	QYERY_APIKEY    = "apikey"
	QUERY_PROVIDERS = "providers"
)

func init() {
	http.HandleFunc("/", handle)
	http.HandleFunc("/proxy.json", handle_cached_json)
	http.HandleFunc("/refresh", handle_refresh_json)
}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1><It Works!/h1>")
}

func handle_cached_json(w http.ResponseWriter, r *http.Request) {
	if !checkPermission(w, r) {
		return
	}
	providers := r.URL.Query().Get(QUERY_PROVIDERS)
	if providers == "" {
		providers = "all"
	}

	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, QUERY_PROVIDERS, providers, 0, nil)

	cache := &ProxyList{}
	err := datastore.Get(ctx, key, cache)

	if err != nil || time.Now().Sub(cache.Timestamp) > time.Hour {
		handle_refresh_json(w, r)
		return
	}

	ret, _ := json.Marshal(cache.Proxies)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(ret))
}

func handle_refresh_json(w http.ResponseWriter, r *http.Request) {
	if !checkPermission(w, r) {
		return
	}

	providers := r.URL.Query().Get(QUERY_PROVIDERS)
	if providers == "" {
		providers = "all"
	}
	providerArr := strings.Split(providers, "|")

	loaders := make([]provider.ProxyProvider, 0)

	for _, name := range providerArr {
		if name == "all" {
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

	cache := &ProxyList{
		Timestamp: time.Now(),
		Proxies:   make([]provider.ProxyItem, len(items)),
	}
	for i, item := range items {
		cache.Proxies[i] = *item
	}

	key := datastore.NewKey(ctx, QUERY_PROVIDERS, providers, 0, nil)
	datastore.Put(ctx, key, cache)

	ret, _ := json.Marshal(items)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(ret))
}

func checkPermission(w http.ResponseWriter, r *http.Request) bool {
	if AUTH_ON {
		apikey := r.URL.Query().Get(QYERY_APIKEY) // TODO
		if apikey != API_KEY {
			errJson(w, ACCESS_DENIED)
			return false
		}
	}
	return true
}

type ProxyList struct {
	Timestamp time.Time
	Proxies   []provider.ProxyItem
}
