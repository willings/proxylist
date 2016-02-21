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
	"os"
	"bytes"
	"encoding/xml"
)

const AUTH_ON = false

const API_KEY = "YOUR_SEC_KEY"

const (
	QYERY_APIKEY = "apikey"
	QUERY_PROVIDERS = "providers"
	QUERY_CACHE = "cache"
)

func init() {
	http.HandleFunc("/", handle)
	http.HandleFunc("/proxy.json", handle_cached_json)
	http.HandleFunc("/proxy.xml", handle_cached_json)
}

func handle(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("static/index.html")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}

	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, QUERY_PROVIDERS, "all", 0, nil)

	cache := &ProxyList{}
	err = datastore.Get(ctx, key, cache)

	updateTime := cache.Timestamp.Format("Mon Jan 2 15:04:05 -0700 MST 2006")

	tblBuf := &bytes.Buffer{}
	tblBuf.Write([]byte("<table><tr><td>IP</td><td>Host</td><td>Type</td><td>Anonymous</td></tr>"))
	for _, proxy := range cache.Proxies {
		var proxyType string
		switch proxy.Type {
		case 3:
			proxyType = "HTTPS, HTTP"
		case 2:
			proxyType = "HTTPS"
		case 1:
			proxyType = "HTTP"
		}
		var anoymousType string
		switch proxy.Anonymous {
		case 1:
			anoymousType = "Anonymous"
		case 0:
			anoymousType = "None"
		}
		tr := fmt.Sprintf("<tr><td>%s</td><td>%d</td><td>%s</td><td>%s</td></tr>",
			proxy.Host, proxy.Port, proxyType, anoymousType)
		tblBuf.Write([]byte(tr))
	}
	tblBuf.Write([]byte("</table>"))

	buf := &bytes.Buffer{}
	buf.ReadFrom(file)

	html := string(buf.Bytes())
	html = strings.Replace(html, "$LAST_UPDATED", updateTime, -1)
	html = strings.Replace(html, "$PROXY_LIST", string(tblBuf.Bytes()), -1)

	fmt.Fprint(w, html)
}

func handle_cached_json(w http.ResponseWriter, r *http.Request) {
	if !checkPermission(w, r) {
		return
	}

	paramCache := r.URL.Query().Get(QUERY_CACHE) != "false"
	providers := r.URL.Query().Get(QUERY_PROVIDERS)
	if providers == "" {
		providers = "all"
	}

	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	key := datastore.NewKey(ctx, QUERY_PROVIDERS, providers, 0, nil)

	cache := &ProxyList{}

	if !paramCache {
		var p provider.ProxyProvider
		if providers == "all" {
			p = provider.CreateAllLoader()
		} else {
			p = provider.CreateProvider(providers)
		}

		if p == nil {
			errJson(w, NO_PROVIDER)
			return
		}

		p.SetClient(client)
		items, err := p.Load()

		if err != nil {
			formatError(w, err)
			return
		}

		cache = &ProxyList{
			Timestamp: time.Now(),
			Proxies:   make([]provider.ProxyItem, len(items)),
		}
		for i, item := range items {
			cache.Proxies[i] = *item
		}
		datastore.Put(ctx, key, cache)
	} else {
		datastore.Get(ctx, key, cache)
	}

	if strings.HasSuffix(r.URL.Path, "xml") {
		ret, _ := xml.Marshal(cache.Proxies)
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, "<proxylist>")
		fmt.Fprint(w, string(ret))
		fmt.Fprint(w, "</proxylist>")
	} else {
		ret, _ := json.Marshal(cache.Proxies)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(ret))
	}
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
