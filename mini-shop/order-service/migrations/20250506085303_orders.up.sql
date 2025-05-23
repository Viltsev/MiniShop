CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    userID INTEGER NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'created',
    createdAt TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
