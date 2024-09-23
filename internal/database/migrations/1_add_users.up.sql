CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id    uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    login text NOT NULL UNIQUE,
    hash  text NOT NULL
);
