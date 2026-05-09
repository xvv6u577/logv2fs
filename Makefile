httpserver: 
	go run ./ httpserver

web:
	cd frontend; npm start;

singbox-server:
	go run -tags with_gvisor,with_quic,with_wireguard,with_utls,with_reality_server,with_clash_api,with_v2ray_api,with_grpc ./ singbox

singbox-client:
	sing-box version && sing-box run -c /Users/guestuser/go/src/github/logv2fs/development/singbox/transit-client.json
