# kafka-trigger

A Kubeless _Trigger_ represents an event source the functions can be associated with it. When an event occurs in the event source, Kubeless will ensure that the associated functions are invoked. __Kafka-trigger__ addon to Kubeless adds support for a Kafka streaming platform as trigger to Kubeless. A Kafka topic can be associated with one or more Kubeless functions. Kubeless functions associated with a topic are triggerd as and when messages get pubslished to the topic.

## Getting started

In Kafka trigger [release page](https://github.com/kubeless/kafka-trigger/releases), you can find the manifest to quickly deploy a collection of Kafka and Zookeeper statefulsets. If you have a Kafka cluster already running in the same Kubernetes environment, you can also deploy PubSub function with it. Check out [this tutorial](/docs/use-existing-kafka) for more details how to do that.

If you want to deploy the manifest we provide to deploy Kafka and Zookeeper execute the following command:

```console
$ export RELEASE=$(curl -s https://api.github.com/repos/kubeless/kafka-trigger/releases/latest | grep tag_name | cut -d '"' -f 4)
$ kubectl create -f https://github.com/kubeless/kakfa-trigger/releases/download/$RELEASE/kafka-zookeeper-$RELEASE.yaml
```

> NOTE: Kafka statefulset uses a PVC (persistent volume claim). Depending on the configuration of your cluster you may need to provision a PV (Persistent Volume) that matches the PVC or configure dynamic storage provisioning. Otherwise Kafka pod will fail to get scheduled. Also note that Kafka is only required for PubSub functions, you can still use http triggered functions. Please refer to [PV](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) documentation on how to provision storage for PVC.

Once deployed, you can verify two statefulsets up and running:

```
$ kubectl -n kubeless get statefulset
NAME      DESIRED   CURRENT   AGE
kafka     1         1         40s
zoo       1         1         42s

$ kubectl -n kubeless get svc
NAME        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
broker      ClusterIP   None            <none>        9092/TCP            1m
kafka       ClusterIP   10.55.250.89    <none>        9092/TCP            1m
zoo         ClusterIP   None            <none>        9092/TCP,3888/TCP   1m
zookeeper   ClusterIP   10.55.249.102   <none>        2181/TCP            1m
```

A function can be as simple as:

```python
def foobar(event, context):
  print event['data']
  return event['data']
```

Now you can deploy a pubsub function. 

```console
$ kubeless function deploy test --runtime python2.7 \
                                --handler test.foobar \
                                --from-file test.py
```

You need to create a _Kafka_ trigger that lets you associate a function with a topic specified by `--trigger-topic` as below:

```console
$ kubeless trigger kafka create test --function-selector created-by=kubeless,function=test --trigger-topic test-topic
```

After that you can invoke the function by publishing messages in that topic. To allow you to easily manage topics `kubeless` provides a convenience function `kubeless topic`. You can create/delete and publish to a topic easily.

```console
$ kubeless topic create test-topic
$ kubeless topic publish --topic test-topic --data "Hello World!"
```

You can check the result in the pod logs:

```console
$ kubectl logs test-695251588-cxwmc
...
Hello World!
```

## Other commands

You can create, list and delete PubSub topics (for Kafka):

```console
$ kubeless topic create another-topic
Created topic "another-topic".

$ kubeless topic delete another-topic

$ kubeless topic ls
```
