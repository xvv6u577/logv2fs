install:
	go build; go install;

backend: 
	go run ./main.go

certs:
	./setup-script-w8/generateCert.sh

v2ray-local:
	./config/v2ray-macos-v4.23.4/v2ray -config ./config/local/config.json

web:
	cd frontend; npm start;

link:
	ln -s /Users/guestuser/go/src/github/logv2rayfullstack/config/nginx/grpc_nginx_80.conf /opt/homebrew/etc/nginx/servers;
	nginx -s reload;

nginx-senario-3:
	ln -s /Users/guestuser/go/src/github/logv2rayfullstack/config/nginx/senario-3.conf /opt/homebrew/etc/nginx/servers;
	nginx -s reload;

nginx-senario-4:
	ln -s /Users/guestuser/go/src/github/logv2rayfullstack/config/nginx/senario-4.conf /opt/homebrew/etc/nginx/servers;
	nginx -s reload;

nginx-senario-5:
	ln -s /Users/guestuser/go/src/github/logv2rayfullstack/config/nginx/senario-5.conf /opt/homebrew/etc/nginx/servers;
	nginx -s reload;

server:
	go run ./cmd/grpcserver/main.go -address 0.0.0.0:50051 

server-tls:
	go run ./cmd/grpcserver/main.go -address 0.0.0.0:50051 -tls

server-tls-auth:
	go run ./cmd/grpcserver/main.go -address 0.0.0.0:50051 -tls -auth

client:
	go run ./cmd/grpcclient/main.go -address 0.0.0.0:50051

client-tls:
	go run ./cmd/grpcclient/main.go -address 0.0.0.0:50051 -tls

client-nginx-tls:
	go run ./cmd/grpcclient/main.go -address 0.0.0.0:8070 -tls
