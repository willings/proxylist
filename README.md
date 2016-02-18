# proxylist

[![Build Status](https://travis-ci.org/willings/proxylist.svg?branch=master)](https://travis-ci.org/willings/proxylist)

A HTTP Proxy list updating service for Google App Engine.

DEMO
------------

http://free-proxy-list.appspot.com/proxy.json

```
{"Host":"200.35.187.114","Port":8080,"Type":3,"Anonymous":1}
{"Host":"83.128.112.158","Port":80,"Type":1,"Anonymous":0}...
```

Usage
------------

```
$  /YOUR_APP_ENGINE_PATH/appcfg.py update proxylist
```