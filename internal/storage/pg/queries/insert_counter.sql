INSERT INTO counters (name, value)
VALUES ($1, $2)
ON CONFLICT (name)
DO UPDATE SET value = counters.value + EXCLUDED.value;
