-- Enable PostGIS extension for geospatial functionality
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable trigram matching for text search
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Verify PostGIS installation
SELECT PostGIS_Version();