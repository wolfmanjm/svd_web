-- name: ListMPUs :many
SELECT *
FROM mpus
ORDER BY name;

-- name: FindMCU :one
SELECT *
FROM mpus
WHERE lower(name) LIKE lower(@name::text);

-- name: FetchPeripherals :many
SELECT *
FROM peripherals WHERE mpu_id = $1
ORDER BY name;

-- name: FindPeripheral :one
SELECT *
FROM peripherals
WHERE mpu_id = $1 AND lower(name) LIKE lower(@name::text);

-- name: FetchRegisters :many
SELECT *
FROM registers WHERE peripheral_id = $1
ORDER BY name;

-- name: FindRegister :one
SELECT *
FROM registers
WHERE peripheral_id = $1 AND lower(name) LIKE lower(@name::text);

-- name: FetchFields :many
SELECT *
FROM fields WHERE register_id = $1
ORDER BY bit_offset;


