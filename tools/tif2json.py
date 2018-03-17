#!/usr/bin/python

"""
tif2json conversion tool
takes a file as TIF
outputs a file as JSON formatted for UNITY
"""

import sys
import os
import subprocess
from os.path import basename
import yaml

def convert(tif):

    path = os.path.dirname(os.path.realpath(tif))
    filename = tif.split('.')
    xyz = filename[0] + ".xyz"
    json = filename[0] + ".json"
    yml = filename[0] + ".yml"
    tifnew = filename[0] + "_3857.tif"
    filebase = basename(filename[0])

    config = yaml.safe_load(open(yml))
    srs = 'EPSG:' + str(config[filebase]['srs'])

    try:

        def warp():
            localoptions = [
                'gdalwarp',
                '--config',
                'GDAL_CACHEMAX',
                '1000',
                '-wm',
                '1000',
                '-overwrite',
                '-s_srs',
                srs,
                '-t_srs',
                'EPSG:3857',
                tif,
                tifnew]
            subprocess.check_output(localoptions, stderr=subprocess.STDOUT)

        def toXYZ():
	    localoptions = [
	        'gdal_translate',
	        '-of',
	        'XYZ',
	        '-a_nodata',
	        '0',
	        tifnew,
	        xyz]
	    subprocess.check_output(localoptions, stderr=subprocess.STDOUT)

        warp()
        toXYZ()
        os.remove(tifnew)

        xyzfile = open(xyz,'r')
        jsonfile = open(json, 'w')

        data = '{"points": ['

        for line in xyzfile:
            values = line.split()

            if values[2] == '0': continue

            tmp = '{'
            tmp += '"x": ' + str(round(float(values[0]),2)) + ','
            tmp += '"y": ' + str(round(float(values[1]),2)) + ','
            tmp += '"z": ' + str(round(float(values[2]),2)-28)
            tmp += '},'

            tmp = tmp.replace("'", '"')
            data += tmp

        data = data[:-1]
        data += ']}'
        
        jsonfile.write(data)

        os.remove(xyz)

    except Exception as e:
        print (str(e))


if (sys.argv[1]):
    convert(sys.argv[1])
else:
    print ("please include a /full/path/name.tif as an argument variable for this tool")
