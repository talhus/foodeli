-- Create admin account for Foodeli
INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
VALUES (
    'a9000000-0000-0000-0000-000000000001',
    'admin@foodeli.fi',
    '$2a$10$U8ardGSXFHlTFu/1AbyCkeIP1xFO4hxcbJOdBvtUY4rh9HjHfcOIu',
    'admin',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO UPDATE SET
    role = 'admin',
    password_hash = '$2a$10$U8ardGSXFHlTFu/1AbyCkeIP1xFO4hxcbJOdBvtUY4rh9HjHfcOIu',
    updated_at = NOW();
