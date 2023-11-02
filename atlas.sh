#!/bin/zsh

../shrugged/shrugged ./asset-source/charset
mv ./charset.* ./assets/

../shrugged/shrugged ./asset-source/bigtiles/entities
mv ./entities.* ./assets/

../shrugged/shrugged ./asset-source/bigtiles/world
mv ./world.* ./assets/