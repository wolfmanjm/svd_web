-- name: AddMCU :one
INSERT INTO mpus (name, description)
VALUES ($1, $2) RETURNING id;

-- name: AddPeripheral :one
INSERT INTO peripherals (mpu_id, name, derived_from_id, base_address, description)
VALUES ($1, $2, $3, $4, $5) RETURNING id;

-- name: AddRegister :one
INSERT INTO registers (peripheral_id, name, address_offset, reset_value, access, description)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;

-- name: AddField :one
INSERT INTO fields (register_id, name, num_bits, bit_offset, access, description)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;

-- name: AddEnumeration :one
INSERT INTO enumerations (field_id, name, value, description)
VALUES ($1, $2, $3, $4) RETURNING id;
