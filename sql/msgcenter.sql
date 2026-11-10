CREATE TABLE `t_msg_template` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `template_id` VARCHAR(256) NOT NULL,
    `rel_template_id` VARCHAR(256) NOT NULL,
    `name` VARCHAR(256) NOT NULL,
    `sign_name` VARCHAR(256),
    `source_id` VARCHAR(256) NOT NULL,
    `channel` INT,
    `subject` VARCHAR(256) NOT NULL,
    `content` VARCHAR(4096) NOT NULL,
    `status` INT,
    `ext` VARCHAR(256),
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `modify_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_msgid` (`template_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;