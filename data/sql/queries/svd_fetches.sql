-- name: ListMPUs :many
SELECT * FROM mpus ORDER BY name;

-- name: FetchPeripherals :many
SELECT id, derived_from_id, name, base_address, description
FROM peripherals WHERE mpu_id = $1
ORDER BY name;
