CREATE TABLE xlsxfiles (
    id SERIAL PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid(),
    "statusCode" VARCHAR(20),
    bucket       VARCHAR(50),
    "originalName" VARCHAR(255),
    "originalSize"         INT,
    "createdAt"    TIMESTAMP DEFAULT NOW(),
    "updatedAt"    TIMESTAMP DEFAULT NOW()
);
