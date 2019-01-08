# kafka-trigger

A Kubeless _Trigger_ represents an event source that functions can be associated with. When an event occurs in the event source, Kubeless will ensure that the associated functions are invoked. __Kafka-trigger__ addon to Kubeless adds support for a Kafka streaming platform as a trigger to Kubeless. A Kafka topic can be associated with one or more Kubeless functions. Kubeless functions associated with a topic are triggered as, and when, messages get published to the topic.

Please refer to the [documentation](https://kubeless.io/docs/pubsub-functions/#kafka) on how to use the Kafka trigger with Kubeless.
