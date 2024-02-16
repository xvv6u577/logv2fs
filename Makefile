backend: 
	go run ./main.go server

web:
	cd frontend; npm start;

install:
	go build; go install;

certs:
	./setup-script-w8/generateCert.sh

local-v2ray:
	./development/v2ray-macos-v4.23.4/v2ray -config ./development/v2ray/local.json

link:
	ln -s /Users/guestuser/go/src/github/logv2fs/development/nginx/grpc_nginx_80.conf /opt/homebrew/etc/nginx/servers;
	nginx -s reload;

nginx-senario-3:
	ln -s /Users/guestuser/go/src/github/logv2fs/development/nginx/senario-3.conf /opt/homebrew/etc/nginx/servers;
	nginx -s reload;

nginx-senario-4:
	ln -s /Users/guestuser/go/src/github/logv2fs/development/nginx/senario-4.conf /opt/homebrew/etc/nginx/servers;
	nginx -s reload;

nginx-senario-5:
	ln -s /Users/guestuser/go/src/github/logv2fs/development/nginx/senario-5.conf /opt/homebrew/etc/nginx/servers;
	nginx -s reload;

server:
	go run ./ GRPCServer --address 0.0.0.0:50051 

server-tls:
	go run ./ GRPCServer --address 0.0.0.0:50051 --tls

server-tls-auth:
	go run ./ GRPCServer --address 0.0.0.0:50051 --tls --auth

client:
	go run ./ GRPCClient --address 0.0.0.0:50051

client-tls:
	go run ./ GRPCClient --address 0.0.0.0:50051 --tls

client-nginx-tls:
	go run ./ GRPCClient --address 0.0.0.0:8070 --tls
