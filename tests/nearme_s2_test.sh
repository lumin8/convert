#!/bin/bash

curl "http://map.life:8000/nearme?s2=5345452ba89&f=json&type=trails" -o trails.json
curl "http://map.life:8000/nearme?s2=5345452ba89&f=json&type=roads" -o roads.json
curl "http://map.life:8000/nearme?s2=5345452ba89&f=json&type=poi" -o pois.json
#curl "http://map.life:8000/dem?s2=5345452ba89" -o dem.json
curl "http://map.life:8000/nearme?s2=5345452ba89&f=json&type=rivers" -o rivers.json
