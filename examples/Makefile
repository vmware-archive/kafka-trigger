pubsub-python:
	kubeless topic create s3-python || true
	kubeless function deploy pubsub-python  --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py
	kubeless trigger kafka create pubsub-python --function-selector created-by=kubeless,function=pubsub-python --trigger-topic s3-python

# Generate a random string to inject into s3 topic,
# then "tail -f" until it shows (with timeout)
pubsub-python-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python --data '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-python`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found
	# Verify event context
	logs=`kubectl logs -l function=pubsub-python`; \
	echo $$logs | grep -q "event-time.*UTC" && \
	echo $$logs | grep -q "event-type.*application/json" && \
	echo $$logs | grep -q "event-namespace.*kafkatriggers.kubeless.io" && \
	echo $$logs | grep -q "event-id.*"

kafka-python-func1-topic-s3-python:
	kubeless topic create s3-python || true
	kubeless function deploy kafka-python-func1-topic-s3-python --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py --label topic=s3-python

kafka-python-func1-topic-s3-python-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python --data '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=kafka-python-func1-topic-s3-python`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
		if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

kafka-python-func2-topic-s3-python:
	kubeless topic create s3-python || true
	kubeless function deploy kafka-python-func2-topic-s3-python --runtime python2.7 --handler pubsub.handler --from-file python/hellowithdata.py --label topic=s3-python

kafka-python-func2-topic-s3-python-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python --data '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=kafka-python-func2-topic-s3-python`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
		if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

s3-python-kafka-trigger:
	kubeless trigger kafka create s3-python-kafka-trigger --function-selector created-by=kubeless,topic=s3-python --trigger-topic s3-python

pubsub-python34:
	kubeless topic create s3-python34 || true
	kubeless function deploy pubsub-python34 --runtime python3.4 --handler pubsub-python.handler --from-file python/hellowithdata34.py
	kubeless trigger kafka create pubsub-python34 --function-selector created-by=kubeless,function=pubsub-python34 --trigger-topic s3-python34

pubsub-python34-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python34 --data '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-python34`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

pubsub-python36:
	kubeless topic create s3-python36 || true
	kubeless function deploy pubsub-python36 --runtime python3.6 --handler pubsub-python.handler --from-file python/pubsub.py
	kubeless trigger kafka create pubsub-python36 --function-selector created-by=kubeless,function=pubsub-python36 --trigger-topic s3-python36

pubsub-python36-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-python36 --data '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-python36`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

pubsub-nodejs:
	kubeless topic create s3-nodejs || true
	kubeless function deploy pubsub-nodejs --runtime nodejs12 --handler pubsub-nodejs.handler --from-file nodejs/hellowithdata.js
	kubeless trigger kafka create pubsub-nodejs --function-selector created-by=kubeless,function=pubsub-nodejs --trigger-topic s3-nodejs

pubsub-nodejs-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-nodejs --data '{"test": "$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-nodejs`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

pubsub-nodejs-update:
	kubeless topic create s3-nodejs-2 || true
	kubeless trigger kafka update pubsub-nodejs --trigger-topic s3-nodejs-2

pubsub-nodejs-update-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-nodejs-2  --data '{"test": "$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-nodejs`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
        if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

pubsub-ruby:
	kubeless topic create s3-ruby || true
	kubeless function deploy pubsub-ruby --runtime ruby2.4 --handler pubsub-ruby.handler --from-file ruby/hellowithdata.rb
	kubeless trigger kafka create pubsub-ruby --function-selector created-by=kubeless,function=pubsub-ruby --trigger-topic s3-ruby

pubsub-ruby-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-ruby --data '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-ruby`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

pubsub-go:
	kubeless topic create s3-go || true
	kubeless function deploy pubsub-go --runtime go1.14 --handler pubsub-go.Handler --from-file golang/hellowithdata.go
	kubeless trigger kafka create pubsub-go --function-selector created-by=kubeless,function=pubsub-go --trigger-topic s3-go

pubsub-go-verify:
	$(eval DATA := $(shell mktemp -u -t XXXXXXXX))
	kubeless topic publish --topic s3-go --data '{"payload":"$(DATA)"}'
	number="1"; \
	timeout="60"; \
	found=false; \
	while [ $$number -le $$timeout ] ; do \
		pod=`kubectl get po -oname -l function=pubsub-go`; \
		logs=`kubectl logs $$pod | grep $(DATA)`; \
    	if [ "$$logs" != "" ]; then \
			found=true; \
			break; \
		fi; \
		sleep 1; \
		number=`expr $$number + 1`; \
	done; \
	$$found

