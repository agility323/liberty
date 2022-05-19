#!/bin/sh

# env var
if [ "$LOG_LEVEL" = "" ]; then
	LOG_LEVEL=1
	echo "env LOG_LEVEL is null, set to $LOG_LEVEL"
fi
if [ "$HOST" = "" ]; then
	HOST=101
	echo "env HOST is null, set to $HOST"
fi
if [ "$CLIENT_SERVER_ADDR" = "" ]; then
	CLIENT_SERVER_ADDR="127.0.0.1:4001"
	echo "env CLIENT_SERVER_ADDR is null, set to $CLIENT_SERVER_ADDR"
fi
if [ "$CONNECT_SERVER_HANDLER" = "" ]; then
	CONNECT_SERVER_HANDLER="{\"service\": \"login_service\", \"method\": \"connect_server\"}"
	echo "env CONNECT_SERVER_HANDLER is null, set to $CONNECT_SERVER_HANDLER"
fi
if [ "$ETCD" = "" ]; then
	ETCD="[\"http://127.0.0.1:2379\", \"http://127.0.0.1:2479\", \"http://127.0.0.1:2579\"]"
	echo "env ETCD is null, set to $ETCD"
fi

# config
cat << EOF > config.json
{
	"log_level": $LOG_LEVEL,
	"host": $HOST,
	"client_server_addr": "$CLIENT_SERVER_ADDR",
	"connect_server_handler": $CONNECT_SERVER_HANDLER,
	"etcd": $ETCD
}
EOF
echo "create config.json done:"
cat config.json

# exec
exec ./gate --conf=config.json
