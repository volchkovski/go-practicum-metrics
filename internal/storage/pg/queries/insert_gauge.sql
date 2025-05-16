INSERT INTO gauges (name, value)
VALUES ($1, $2)
ON CONFLICT (name)
DO UPDATE SET value = EXCLUDED.value;
