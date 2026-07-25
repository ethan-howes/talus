#!/usr/bin/env python3

import sys
import numpy as np
from scipy import stats
from osgeo import gdal

gdal.UseExceptions()

NODATA = -9999.0
PASS_THRESHOLD = 0.99

if len(sys.argv) != 3:
    print("Usage: validate_output.py <cuda_output.bin> <qgis_reference.tif>")
    sys.exit(1)

cuda_path = sys.argv[1]
qgis_path = sys.argv[2]

cuda_data = np.fromfile(cuda_path, dtype=np.float32)

ds = gdal.Open(qgis_path)
if ds is None:
    print(f"ERROR: could not open {qgis_path}")
    sys.exit(1)
qgis_data = ds.GetRasterBand(1).ReadAsArray().flatten().astype(np.float32)
ds = None

if cuda_data.shape != qgis_data.shape:
    print(f"ERROR: size mismatch: CUDA {cuda_data.shape} vs QGIS {qgis_data.shape}")
    sys.exit(1)

valid = (cuda_data != NODATA) & (qgis_data != NODATA) & \
        np.isfinite(cuda_data) & np.isfinite(qgis_data)

cuda_valid = cuda_data[valid]
qgis_valid = qgis_data[valid]

print(f"Valid cells: {valid.sum():,} of {len(cuda_data):,}")

r, p = stats.pearsonr(cuda_valid, qgis_valid)

print(f"Pearson r:  {r:.6f}")
print(f"p-value:    {p:.2e}")
print(f"Threshold:  {PASS_THRESHOLD}")
print()

if r >= PASS_THRESHOLD:
    print(f"PASS: correlation {r:.4f} exceeds {PASS_THRESHOLD}")
    sys.exit(0)
else:
    print(f"FAIL: correlation {r:.4f} below {PASS_THRESHOLD}")
    sys.exit(1)
