CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE transactions
(
    id           BIGINT PRIMARY KEY AUTO_INCREMENT,
    from_user_id INT            NOT NULL,
    to_user_id   INT            NOT NULL,
    amount       DECIMAL(15, 2) NOT NULL,
    transaction_type         VARCHAR(30)    NOT NULL,
    status       VARCHAR(30)    NOT NULL,
    created_at   TIMESTAMP      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE balances (
    user_id BIGINT PRIMARY KEY,
    amount DECIMAL(15,2) NOT NULL,
    last_updated_at TIMESTAMP
);

CREATE TABLE audit_logs (
   id BIGINT PRIMARY KEY AUTO_INCREMENT,
   entity_type VARCHAR(100) NOT NULL,
   entity_id BIGINT NOT NULL,
   action VARCHAR(50) NOT NULL,
   details TEXT,
   created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);