# Homebroker-go 

## Initial configs

insert in /etc/hosts, this line: 

```shell
# docker internal
127.0.0.1     kubernetes.docker.internal host.docker.internal
127.0.0.1     kubernetes.docker.internal
127.0.0.1     host.docker.internal
172.17.0.1    kafka
# end section
```

init application:

```shell
docker-compose up -d
```

access Confluent Control Center:

```shell
http://localhost:9021/clusters
```

create two topics for Kafka: input/output

json test input: 
```json
{
   "order_id":"1",
   "investor_id":"Mari",
   "asset_id":"asset1",
   "current_shares":10,
   "shares":5,
   "price":5.0,
   "order_type":"SELL"
}
```