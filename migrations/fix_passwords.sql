UPDATE users
SET password_hash = '$2a$10$U8ardGSXFHlTFu/1AbyCkeIP1xFO4hxcbJOdBvtUY4rh9HjHfcOIu'
WHERE email IN (
    'customer@foodeli.fi',
    'merchant@foodeli.fi',
    'kuurna.bistro@foodeli.fi',
    'friends.brgrs@foodeli.fi',
    'tokyo.ramen@foodeli.fi',
    'green.hippo@foodeli.fi'
);
