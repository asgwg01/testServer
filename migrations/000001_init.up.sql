CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL
);

CREATE TABLE IF NOT EXISTS services (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL,
    price INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS subscriptions (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_id UUID NOT NULL,
    user_id UUID NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL
);
