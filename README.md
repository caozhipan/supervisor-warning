supervisor
XMLRPC interface
包依赖：
1:github.com/kolo/xmlrpc
go get github.com/kolo/xmlrpc

2:gopkg.in/gomail.v1
go get gopkg.in/gomail.v1

要求开启supervisord.conf中
[inet_http_server]         ; inet (TCP) server disabled by default
port=*:9001
