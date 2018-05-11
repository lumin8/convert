curl "http://map.life:8000/nearme?x=-111.02593&y=45.63856&f=json&type=trails" -o trails.json
curl "http://map.life:8000/nearme?x=-111.02593&y=45.63856&f=json&type=roads" -o roads.json
curl "http://map.life:8000/nearme?x=-111.02593&y=45.63856&f=json&type=poi" -o pois.json
curl "http://map.life:8000/dem?x=-111.02593&y=45.63856" -o dem.json
curl "http://map.life:8000/nearme?x=-111.02593&y=45.63856&f=json&type=rivers" -o rivers.json
