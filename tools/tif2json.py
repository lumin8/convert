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

#os.chdir(os.path.dirname(sys.argv[0]))

def convert(tif):

    path = os.path.dirname(os.path.realpath(tif))
    filename = tif.split('.')
    xyz = filename[0] + ".xyz"
    json = filename[0] + ".json"

    try:
        def toXYZ():
	    localoptions = [
	        'gdal_translate',
	        '-of',
	        'XYZ',
	        '-a_nodata',
	        '0',
	        tif,
	        xyz]
	    subprocess.check_output(localoptions, stderr=subprocess.STDOUT)

        toXYZ()

        xyzfile = open(xyz,'r')
        jsonfile = open(json, 'w')

        data = '{"points": ['

        for line in xyzfile:
            values = line.split()

            if values[2] == '0': continue

            data += '{'
            data += '"x": ' + values[0] + ','
            data += '"y": ' + values[1] + ','
            data += '"z": ' + values[2]
            data += '},'

        data = data[:-1]
        data += ']}'
        
        jsonfile.write(data)

        os.remove(xyz)

    except Exception as e:
        print (str(e))


if (sys.argv[1]):
    convert(sys.argv[1])
else:
    print ("please include a filepath/name as an argument variable for this tool")
