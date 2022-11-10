
// Simulation DB
TRUNCATE config_data,config_metadata,config_master;
ALTER TABLE config_master ADD scope VARCHAR NOT NULL;

// Product DB
ALTER TABLE products_equipments ADD allocated_metric VARCHAR NOT NULL DEFAULT '';
ALTER TABLE overall_computed_licences ADD cost_optimization BOOLEAN DEFAULT FALSE;