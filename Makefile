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

# always keep!
encrypt:
	echo "hello world" | openssl enc -aes-256-cbc -pbkdf2 -salt -base64

decrypt:
	echo "U2FsdGVkX1+vH21Ft9rzScabAsa7BJw6nHRPDpRJC0A=" | openssl enc -d -aes-256-cbc -pbkdf2 -base64

makefile-test:
	echo "U2FsdGVkX1+faoSxUmKZ1XUgzpgjWXIM7TsyVjg2GoW//apZbiVzIWkPCO+U4XYB9plByo8pBVChKZYhAfU9C1vOaIQOf96fqRUlg3SUBv+qYxNWCVYlor+wBiBkGqJCllfllGdELrSX7QvOQCTe7XxWrVNYNdUtqwHWkpw9W2nNRSs5q9TzuXQzz9sOSe6x" | openssl enc -d -aes-256-cbc -pbkdf2 -base64