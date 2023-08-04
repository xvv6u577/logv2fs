#### Senario 1 (GRPC client noTLS <-> GRPC server) Success!

<!-- Comment client AddUser at first -->
```bash
make server
make client
```

#### Senario 2 (GRPC client <-> TLS GRPC server) Success!

<!-- Comment client AddUser at first -->
```bash
make server-tls
make client-tls
```

#### Senario 3 (GRPC client <-> TLS <-> Nginx <-> noTLS <-> GRPC server) Success!

```bash
make nginx-senario-3
make server
make client-nginx-tls
```

#### Senario 4 (GRPC client <-> TLS <-> nginx <-> TLS(no auth) <-> GRPC server) Success!

```bash
make nginx-senario-4
make server-tls
make client-nginx-tls
```

#### Senario 5 (GRPC client <-> TLS <-> nginx <-> TLS(with auth) <-> GRPC server) Failed!

```bash
make nginx-senario-5
make server-tls-auth
make client-nginx-tls
```