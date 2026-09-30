# apache artemis v2.54.0
console username: `artemis-activemq`

console password: `artemis-activemq`

console URL: http://localhost:8161

## URL
`amqp://localhost:5672`
## address
`some.svc.group`

## multicast (or topic)
`topic.name`

## FQQN
`some.svc.group::topic.name`

## add user
```
/var/lib/artemis-test-cluster/bin/artemis user add \
  --user-command-user svc-producer \
  --user-command-password test \
  --role svc-producer-role
```
```
/var/lib/artemis-test-cluster/bin/artemis user add \
  --user-command-user svc-consumer \
  --user-command-password test \
  --role svc-consumer-role
```

### producer: username:`svc-producer` password:`test`
### consumer: username:`svc-consumer` password:`test`
### consumer: `svc-some_service` :-)

# login attempts topic settings (artemis)
```xml
         <!-- some svc group -->
         <security-setting match="some.svc.group#">
            <permission type="manage" roles="amq"/>
            <permission type="createNonDurableQueue" roles="amq"/>
            <permission type="deleteNonDurableQueue" roles="amq"/>
            <permission type="createDurableQueue" roles="amq"/>
            <permission type="deleteDurableQueue" roles="amq"/>
            <permission type="createAddress" roles="amq"/>
            <permission type="deleteAddress" roles="amq"/>
            <permission type="consume" roles="amq,svc-consumer-role"/>
            <permission type="browse" roles="amq,svc-consumer-role,svc-producer-role"/>
            <permission type="send" roles="amq,svc-producer-role"/>
         </security-setting>
```
```xml
         <!-- some.svc.group -->
         <address-setting match="some.svc.group::topic.name">
            <dead-letter-address>some.svc.group.DLQ::topic.name.DLQ</dead-letter-address>
            <expiry-address>some.svc.group.Expiry::topic.name.Expiry</expiry-address>
         </address-setting>
```
```xml
         <!-- some svc group -->
         <address name="some.svc.group.DLQ">
            <anycast>
               <queue name="some.svc.group.DLQ" />
            </anycast>
         </address>
         <address name="some.svc.group.Expiry">
            <anycast>
               <queue name="some.svc.group.Expiry" />
            </anycast>
         </address>

         <address name="some.svc.group">
            <multicast>
               <queue name="topic.name" />
            </multicast>
         </address>
```

## send test message (artemis cli)
`producer --destination topic://some.svc.group::topic.name --message "test" --message-count 1`

## browse messages (artemis cli)
`browser --destination some.svc.group::topic.name`

## consume message (artemis cli)
`consumer --user=svc-consumer --password=test --destination topic://some.svc.group::topic.name --message-count 1`

