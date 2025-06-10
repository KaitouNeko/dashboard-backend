-- db/migrations/000X_add_humidity_to_energy_tables.down.sql
ALTER TABLE energy_usage DROP COLUMN humidity_percent;
ALTER TABLE energy_forecasts DROP COLUMN humidity_percent;
