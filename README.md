# Data
It's all about the data.  CRUD, with conversion into the filetypes necessary for the DeepAR client apps.

## Sample
Several datasets included.

Clients should only Use those in EPSG:3857 (preferrably).

Sample sets in other projections systems (spatial reference systems / EPSG codes) are for conversion testing in the Data api.

Data needs to be
- converted to a standard projection (what does Unity use?)
- converted to dxf or other as needed (what does Unity need?)

At this time, data is just in 3D point structure (x,y,z with attributes).

Datasets for consumption by client also now include a YML describing the dataset.  See sample/somedata.yml for an example.
