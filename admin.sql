CREATE TABLE IF NOT EXISTS robot (
    `robot_id` VARCHAR(32) NOT NULL PRIMARY KEY,
    `owner` VARCHAR(15) NOT NULL,
    `wxid` VARCHAR(32)  NULL,
    `device_id` VARCHAR(32)  NULL,
    `device_name` VARCHAR(32)  NULL,
    `server_host` VARCHAR(32)  NULL,
    `server_port` INT  NULL
);