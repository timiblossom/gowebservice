/* *****************************************************************************
// Setup the preferences
// ****************************************************************************/
SET NAMES utf8 COLLATE 'utf8_unicode_ci';
SET foreign_key_checks = 1;
SET time_zone = '+00:00';
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';
SET default_storage_engine = InnoDB;
SET CHARACTER SET utf8;

SET GLOBAL sql_mode="STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION";
SET SESSION sql_mode="STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION";

/* *****************************************************************************
// Remove old database
// ****************************************************************************/
DROP DATABASE IF EXISTS puma;

/* *****************************************************************************
// Create new database
// ****************************************************************************/
CREATE DATABASE puma DEFAULT CHARSET = utf8 COLLATE = utf8_unicode_ci;
USE puma;

/* *****************************************************************************
// Create the tables
// ****************************************************************************/
CREATE TABLE user_status (
    id TINYINT(1) UNSIGNED NOT NULL AUTO_INCREMENT,
    status VARCHAR(25) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT 0,
    
    PRIMARY KEY (id)
);

CREATE TABLE user_role (
    id INT(10) NOT NULL AUTO_INCREMENT,
    role_name varchar(50) NOT NULL,
    description varchar(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT 0,

    PRIMARY KEY (id)
);

CREATE TABLE user (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password CHAR(60) NOT NULL,
    
    status_id TINYINT(1) UNSIGNED NOT NULL DEFAULT 1,
    user_role INT(10) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted TINYINT(1) UNSIGNED NOT NULL DEFAULT 0,
    
    UNIQUE KEY (email),
    CONSTRAINT `f_user_status` FOREIGN KEY (`status_id`) REFERENCES `user_status` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `f_user_role` FOREIGN KEY (`user_role`) REFERENCES `user_role` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
    
    PRIMARY KEY (id)
);

