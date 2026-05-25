httpserver: 
	cd /Users/guestuser/go/src/github/logv2fs;go run ./ httpserver

web:
	cd /Users/guestuser/go/src/github/logv2fs/frontend; npm start;

singbox-server:
	cd /Users/guestuser/go/src/github/logv2fs;go run -tags with_gvisor,with_quic,with_wireguard,with_utls,with_reality_server,with_clash_api,with_v2ray_api,with_grpc ./ singbox

singbox-server-2nd:
	cd /Users/guestuser/go/src/github/logv2fs;go run -tags with_gvisor,with_quic,with_wireguard,with_utls,with_reality_server,with_clash_api,with_v2ray_api,with_grpc ./ singbox --config /Users/guestuser/go/src/github/logv2fs/development/singbox/transit-server-2nd.json --env /Users/guestuser/go/src/github/logv2fs/.env_2nd

singbox-client:
	cd /Users/guestuser/go/src/github/logv2fs;sing-box version && sing-box run -c /Users/guestuser/go/src/github/logv2fs/development/singbox/transit-client.json
