CREATE TABLE `transactions` (
    `id` BIGINT(64) NOT NULL,
    `date` DATETIME NOT NULL,
    `account` VARCHAR(64) NOT NULL,
    `category` VARCHAR(64) NOT NULL,
    `amount` DOUBLE(64, 2) NOT NULL,
    `note` TEXT NOT NULL,
    `type` VARCHAR(16) NOT NULL,
    `position` JSON DEFAULT NULL,
    `created_at` DATETIME NOT NULL,
    `created_by` VARCHAR(64) DEFAULT NULL,
    `updated_at` DATETIME DEFAULT NULL,
    `updated_by` VARCHAR(64) DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;