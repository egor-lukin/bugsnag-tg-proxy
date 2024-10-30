# Bugsnag Telegram Proxy

## Test

``` sh
curl -v localhost:8080/webhooks -X POST --header "Content-Type: application/json" -d @test.json
```

## Deploy 

- docker

``` sh
docker pull ghcr.io/egor-lukin/bugsnag-tg-proxy:721f2716593527e39f359c3f7a1aa59e3e29fd36
```

- helm

``` sh
helm upgrade --install -n bugsnag-proxy bugsnag-proxy .
```
