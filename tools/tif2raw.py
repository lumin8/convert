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
import glob

def convert(tif):

    path = os.path.dirname(os.path.realpath(tif))
    filename = tif.split('.')
    raw = filename[0] + ".raw"
    header = filename[0] + ".hdr"
    tifnew = filename[0] + "_3857.tif"
    yml = filename[0] + ".yml"
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
        
        def toraw():
            os.chdir(path)
	    options = [
	        'gdal_translate',
	        '-ot',
	        'UInt16',
	        '-scale',
	        '-of',
                'ENVI',
                '-outsize',
                '2049',
                '2049',
                '-a_nodata',
                '0',
	        tifnew,
	        raw]
	    subprocess.check_output(options)

        searchpath = path + "/" + filebase + ".raw*"
        for name in glob.glob(searchpath):
            try:
                os.remove(name)
                os.remove(hdr)
            except OSError:
                pass

        warp()
        toraw()
        os.remove(tifnew)

    except Exception as e:
        print (str(e))


if (sys.argv[1]):
    convert(sys.argv[1])
else:
    print ("please include a filepath/name as an argument variable for this tool")
