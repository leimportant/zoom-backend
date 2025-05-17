# Zoom Backend API

Backend API untuk aplikasi Zoom meeting management menggunakan Go dan MySQL.

## Prasyarat

- Go 1.18+ sudah terinstall di komputer Anda
- MySQL server berjalan dan siap digunakan

## Clone Repository

```bash
git clone https://github.com/leimportant/zoom-backend.git
cd zoom-backend

go mod download
go build -o zoom-backend

```
## Buat Database zoomdb
# Create Table
```bash
CREATE TABLE `zoom_meetings` (
  `id` VARCHAR(36) NOT NULL COLLATE 'utf8mb4_unicode_ci',
  `zoom_id` VARCHAR(255) NULL DEFAULT NULL COLLATE 'utf8mb4_unicode_ci',
  `topic` TEXT NULL DEFAULT NULL COLLATE 'utf8mb4_unicode_ci',
  `agenda` TEXT NULL DEFAULT NULL COLLATE 'utf8mb4_unicode_ci',
  `start_time` DATETIME NULL DEFAULT NULL,
  `duration` INT(11) NULL DEFAULT NULL,
  `join_url` TEXT NULL DEFAULT NULL COLLATE 'utf8mb4_unicode_ci',
  `start_url` TEXT NULL DEFAULT NULL COLLATE 'utf8mb4_unicode_ci',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE
)
COLLATE='utf8mb4_unicode_ci'
ENGINE=InnoDB;
```