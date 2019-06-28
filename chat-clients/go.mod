module github.com/fx-wood/dicemagic/chat-clients

go 1.12

require (
	cloud.google.com/go v0.36.0
	contrib.go.opencensus.io/exporter/stackdriver v0.9.1
	github.com/fx-wood/dicemagic v0.0.0-20190306205428-6b9ac5ae3d91
	github.com/fx-wood/dicemagic/internal/dicelang v0.1.0
	github.com/fx-wood/dicemagic/internal/dicelang/errors v0.1.0
	github.com/fx-wood/dicemagic/internal/handler v0.1.0
	github.com/fx-wood/dicemagic/internal/logger v0.1.0
	github.com/fx-wood/dicemagic/internal/proto v0.1.0
	github.com/census-instrumentation/opencensus-proto v0.1.0 // indirect
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/gorilla/mux v1.7.0
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/nlopes/slack v0.5.0
	github.com/serialx/hashring v0.0.0-20180504054112-49a4782e9908
	go.opencensus.io v0.19.1
	golang.org/x/net v0.0.0-20190301231341-16b79f2e4e95
	golang.org/x/oauth2 v0.0.0-20190226205417-e64efc72b421
	google.golang.org/api v0.1.0
	google.golang.org/grpc v1.19.0
)

replace github.com/fx-wood/dicemagic/internal/dicelang v0.1.0 => ../internal/dicelang

replace github.com/fx-wood/dicemagic/internal/dicelang/errors v0.1.0 => ../internal/dicelang/errors

replace github.com/fx-wood/dicemagic/internal/handler v0.1.0 => ../internal/handler

replace github.com/fx-wood/dicemagic/internal/logger v0.1.0 => ../internal/logger

replace github.com/fx-wood/dicemagic/internal/proto v0.1.0 => ../internal/proto
