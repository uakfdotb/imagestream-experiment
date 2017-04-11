CREATE TABLE experiments (
	id VARCHAR(16) NOT NULL PRIMARY KEY,
	total_images INT DEFAULT 1000,
	conventional_redundancy INT DEFAULT 3,
	rapid_redundancy INT DEFAULT 6,
	images_per_task INT DEFAULT 100,
	conventional_tasks_per_worker INT DEFAULT 1,
	rapid_tasks_per_worker INT DEFAULT 2
);

CREATE TABLE images (
	id INT NOT NULL PRIMARY KEY,
	is_in_class TINYINT(1) NOT NULL
);

CREATE TABLE tasks (
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	experiment_id VARCHAR(16) NOT NULL,
	completed TINYINT(1) NOT NULL DEFAULT 0,
	type ENUM('conventional', 'rapid') NOT NULL,
	locked_until TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE task_images (
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	task_id INT NOT NULL,
	image_id INT NOT NULL
);

CREATE TABLE workers (
	id INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
	experiment_id VARCHAR(16) NOT NULL,
	time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE task_results (
	worker_id INT NOT NULL,
	task_id INT NOT NULL,
	duration_ms INT NOT NULL
);

CREATE TABLE conventional_labels (
	worker_id INT NOT NULL,
	task_id INT NOT NULL,
	task_image_id INT NOT NULL,
	is_in_class TINYINT(1) NOT NULL
);

CREATE TABLE rapid_logs (
	worker_id INT NOT NULL,
	task_id INT NOT NULL,
	display_json VARCHAR(8192) NOT NULL,
	click_json VARCHAR(8192) NOT NULL,
	mean DECIMAL(16, 8) NOT NULL,
	sigma DECIMAL(16, 8) NOT NULL
);
