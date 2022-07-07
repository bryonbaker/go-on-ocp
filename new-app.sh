#! /bin/bash

oc new-app golang~https://github.com/sclorg/golang-ex.git
oc expose deployment/golang-ex --port=8888
oc expose svc/golang-ex --hostname=golang-ex-s2i-test.apps-crc.testing
