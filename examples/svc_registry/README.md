# Summary

This is a more realistic example.

Imagine you have a command-line tool to perform service lookups. Some services are more
"service-like" than others, like gRPC, and accept connections. Others consume messages
from a broker, such as Pub/Sub or Kafka.

If a user wants to lookup properties of a service, not all fields will be relevant to
all services. The tab completed suggestions here will adapt to the service that has
been provided.

```
# Will suggest the services themselves due to ValidArgs
svc_registry <TAB>

# Will suggest 'grpc_addr'
svc_registry server1 -f <TAB>

# Will suggest 'pubsub_topic' and 'pubsub_subscription'
svc_registry consumer1 -f <TAB>
```

Read `main_test.go` to see all of the outcomes.

# Testing

To test the completion live, build the binary and 


```
go build -o /tmp/svc_regisry
COMP_INSTALL=1 /tmp/svc_regisry

/tmp/svc_registry <TAB>
/tmp/svc_registry server1 -f <TAB>
/tmp/svc_registry consumer1 -f <TAB>

COMP_UNINSTALL=1 /tmp/svc_regisry
```
