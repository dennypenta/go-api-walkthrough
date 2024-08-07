CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4() NOT NULL,
    username varchar(55) NOT NULL,

    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    deletedAt TIMESTAMP
);
 
CREATE INDEX idx_users_deletedat_createdat ON users (deletedAt, createdAt DESC);
