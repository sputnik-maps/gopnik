#!/bin/bash

cd /gopnik/bin
./gopnikrender --config ../example/dockerconfig.json &
./gopnikdispatcher --config ../example/dockerconfig.json