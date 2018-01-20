#!/usr/bin/python

"""
csv2json conversion tool
takes a file as CSV
outputs a file as JSON formatted for UNITY
"""

import sys
import os
from os.path import basename
import csv
import yaml

def convert(csvf):

    path = os.path.dirname(os.path.realpath(csvf))
    filename = csvf.split('.')
    json = filename[0] + ".json"
    yml = filename[0] + ".yml"
    csvftmp = filename[0] + "_tmp.csv"
    filebase = basename(csvf)

    config = yaml.safe_load(open(yml))
    jsonfile = open(json, 'w')
    
    x = config[filebase]['x_field']
    y = config[filebase]['y_field']
    z = config[filebase]['z_field']
    geom = config[filebase]['geom']

    try:

        #first, map the headers, to a new file
        with open(csvf, 'rb') as infile, open(csvftmp, 'wb') as outfile:
            r = csv.reader(infile)
            w = csv.writer(outfile)

            r_reader = csv.DictReader(infile)

            headers = r_reader.fieldnames

            headers = ["x" if h==x else h for h in headers]
            headers = ["y" if h==y else h for h in headers]
            headers = ["z" if h==z else h for h in headers]

            next(r, None)  # skip the first row from the reader, the old header
            w.writerow(headers)

            for row in r:
                w.writerow(row)

        #second, write the json
        with open(csvftmp, 'rb') as tmp:
            d_reader = csv.DictReader(tmp)

            data = '{"' + geom + '": ['

            for line in d_reader:

                tmp = str(line).replace("'", '"') 
                data += tmp + ','

            data = data[:-1]

            data += ']}'
        
        jsonfile.write(data)

        os.remove(csvftmp)

    except Exception as e:
        print (str(e))


if (sys.argv[1]):
    convert(sys.argv[1])
else:
    print ("please include a /full/path/and/name.csv as an argument variable for this tool")
