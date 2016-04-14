GOPNIK
======
Gopnik is a tile server and a render for slippy map based on mapnik library.
See [sputnik-maps.github.io/gopnik](http://sputnik-maps.github.io/gopnik/)

BUILD
=====

    ./bootstrap.bash
    ./build.bash

BUILD AND RUN USING DOCKER
==========================

* To build an image:

`$ docker build -t gopnik .`

* Run gopnik server with example:

`$ docker run -it --rm -p 8080:8080 -p 9090:9090 gopnik`

To see an example open example/index.html in your favorite browser and zoom out a map to
level 1.

Image is exposing volume /gopnik_data/. You can customise configuration by mounting your volume with
config.json file inside and other helper stuff, e.g. styles. Also, you can use this image as a base
image for your configurations.