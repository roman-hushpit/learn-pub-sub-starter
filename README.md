# learn-pub-sub-starter (Peril)

This is the starter code used in Boot.dev's [Learn Pub/Sub](https://learn.boot.dev/learn-pub-sub) course.


./rabbit.sh start

Go to the RabbitMQ management UI at http://localhost:15672 and navigate to the "Exchanges" tab. Create a new exchange called peril_direct with the type direct.

Create a new exchange with type topic, and name it peril_topic.


Using the UI, create a new exchange called peril_dlx of type fanout Fanout is a good choice because we want all failed messages sent to the exchange to be routed to the queue, without needing to worry about routing keys. Use the default settings.
Using the UI, create a new queue called peril_dlq. Go to the queue's page and bind the queue to the peril_dlx exchange with no routing key. Leave the default settings.