-- ==============================================================================
-- Foodeli Finland Sample Restaurants & Menus Seed Script
-- All names, descriptions, and addresses are localized to Finland and in English.
-- Prices are stored in minor currency units (cents EUR): e.g. 1450 = €14.50.
-- ==============================================================================

BEGIN;

-- 1. Update existing stores with authentic Finland addresses and English names
UPDATE stores
SET name = 'Pizzeria Via Tribunali',
    address = 'Sofiankatu 4, 00170 Helsinki, Finland',
    updated_at = NOW()
WHERE id = '485a6271-37fe-483c-b8ec-7e20365689cd';

UPDATE stores
SET name = 'Nordic Salmon & Seafood Bar',
    address = 'Kauppatori 1, 00170 Helsinki, Finland',
    updated_at = NOW()
WHERE id = '00d19df4-c72e-488d-8087-70b4fb18cc36';

UPDATE stores
SET name = 'Mario''s Artisan Pizza & Pasta',
    address = 'Fredrikinkatu 34, 00120 Helsinki, Finland',
    updated_at = NOW()
WHERE id = '55ea6025-f6af-49b6-92af-7826844fad06';

UPDATE stores
SET name = 'Pamukkale Mediterranean Grill',
    address = 'Lielahdenkatu 14, 33400 Tampere, Finland',
    updated_at = NOW()
WHERE id = '9df1d050-2e8b-4113-87a2-07b52423dff5';

-- 2. Update existing legacy products to have proper descriptions and prices
UPDATE products
SET name = 'Margherita Classica',
    description = 'San Marzano D.O.P. tomatoes, fresh buffalo mozzarella, extra virgin olive oil, and fresh basil on artisan wood-fired crust.',
    price = 1450,
    stock = 50,
    updated_at = NOW()
WHERE id = 'ec56f5d3-d4bd-48c1-9bd5-601fe491fe3f';

UPDATE products
SET name = 'Charcoal Grilled Köfte Platter',
    description = 'Spiced minced lamb & beef patties grilled over wood charcoal, served with roasted sweet peppers, saffron rice, and garlic yogurt.',
    price = 1650,
    stock = 45,
    updated_at = NOW()
WHERE id = '9e99bdfd-639b-4271-ba6f-8484fd033fb8';

-- 3. Insert new merchant users (default password is password123)
INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
VALUES
    ('a1000000-0000-0000-0000-000000000001', 'kuurna.bistro@foodeli.fi', '$2a$10$bvbNPmBpnC2J/lO.b8qMqOLOzz9BeDqxmhQAITwcFuplRIXHHTFx2', 'merchant', NOW(), NOW()),
    ('a2000000-0000-0000-0000-000000000002', 'friends.brgrs@foodeli.fi', '$2a$10$bvbNPmBpnC2J/lO.b8qMqOLOzz9BeDqxmhQAITwcFuplRIXHHTFx2', 'merchant', NOW(), NOW()),
    ('a3000000-0000-0000-0000-000000000003', 'tokyo.ramen@foodeli.fi', '$2a$10$bvbNPmBpnC2J/lO.b8qMqOLOzz9BeDqxmhQAITwcFuplRIXHHTFx2', 'merchant', NOW(), NOW()),
    ('a4000000-0000-0000-0000-000000000004', 'green.hippo@foodeli.fi', '$2a$10$bvbNPmBpnC2J/lO.b8qMqOLOzz9BeDqxmhQAITwcFuplRIXHHTFx2', 'merchant', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- 4. Insert new stores for each merchant
INSERT INTO stores (id, merchant_id, name, address, created_at, updated_at)
VALUES
    ('b1000000-0000-0000-0000-000000000001', 'a1000000-0000-0000-0000-000000000001', 'Ravintola Kuurna & Nordic Bistro', 'Meritullinkatu 6, 00170 Helsinki, Finland', NOW(), NOW()),
    ('b2000000-0000-0000-0000-000000000002', 'a2000000-0000-0000-0000-000000000002', 'Friends & Brgrs Helsinki', 'Mannerheimintie 20, 00100 Helsinki, Finland', NOW(), NOW()),
    ('b3000000-0000-0000-0000-000000000003', 'a3000000-0000-0000-0000-000000000003', 'Tokyo Street & Ramen House', 'Urho Kekkosen katu 1, 00100 Helsinki, Finland', NOW(), NOW()),
    ('b4000000-0000-0000-0000-000000000004', 'a4000000-0000-0000-0000-000000000004', 'Green Hippo Cafe & Bowls', 'Punavuorenkatu 2, 00120 Helsinki, Finland', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    address = EXCLUDED.address,
    updated_at = NOW();

-- 5. Insert Products for Pizzeria Via Tribunali (Store ID: 485a6271-37fe-483c-b8ec-7e20365689cd)
INSERT INTO products (id, store_id, name, description, price, stock, created_at, updated_at)
VALUES
    ('c1010000-0000-0000-0000-000000000001', '485a6271-37fe-483c-b8ec-7e20365689cd', 'Margherita Verace D.O.P.', 'San Marzano tomatoes, buffalo mozzarella from Campania, fresh basil, and extra virgin olive oil.', 1490, 50, NOW(), NOW()),
    ('c1010000-0000-0000-0000-000000000002', '485a6271-37fe-483c-b8ec-7e20365689cd', 'Diavola Piccante', 'Spicy Calabrian spianata salame, San Marzano tomato sauce, fior di latte mozzarella, and chili oil.', 1690, 45, NOW(), NOW()),
    ('c1010000-0000-0000-0000-000000000003', '485a6271-37fe-483c-b8ec-7e20365689cd', 'Tartufo & Forest Mushrooms', 'White truffle cream base, roasted wild portobello mushrooms, smoked provolone, and fresh thyme.', 1850, 35, NOW(), NOW()),
    ('c1010000-0000-0000-0000-000000000004', '485a6271-37fe-483c-b8ec-7e20365689cd', 'Quattro Formaggi & Wild Honey', 'Fior di latte, Gorgonzola D.O.P., aged Parmesan flakes, creamy ricotta, and drizzle of organic wild honey.', 1720, 40, NOW(), NOW()),
    ('c1010000-0000-0000-0000-000000000005', '485a6271-37fe-483c-b8ec-7e20365689cd', 'Traditional Artisan Tiramisu', 'Savoiardi ladyfingers soaked in dark espresso and Marsala wine, layered with velvety mascarpone cream.', 750, 30, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock = EXCLUDED.stock,
    updated_at = NOW();

-- 6. Insert Products for Nordic Salmon & Seafood Bar (Store ID: 00d19df4-c72e-488d-8087-70b4fb18cc36)
INSERT INTO products (id, store_id, name, description, price, stock, created_at, updated_at)
VALUES
    ('c1020000-0000-0000-0000-000000000001', '00d19df4-c72e-488d-8087-70b4fb18cc36', 'Traditional Finnish Salmon Soup (Lohikeitto)', 'Velvety cream broth with fresh Arctic salmon cubes, tender potatoes, leeks, and mountain of aromatic dill. Served with archipelago dark rye bread.', 1580, 60, NOW(), NOW()),
    ('c1020000-0000-0000-0000-000000000002', '00d19df4-c72e-488d-8087-70b4fb18cc36', 'Flame-Grilled Loimulohi Salmon Bowl', 'Open-fire cedar-planked smoked salmon, Finnish pearl barley risotto, pickled red onions, cucumber ribbons, and dill aioli.', 1950, 40, NOW(), NOW()),
    ('c1020000-0000-0000-0000-000000000003', '00d19df4-c72e-488d-8087-70b4fb18cc36', 'Gravlax on Archipelago Toast', 'Beetroot-cured Arctic salmon carpaccio on toasted archipelago dark rye bread with sweet honey-mustard sauce and capers.', 1390, 50, NOW(), NOW()),
    ('c1020000-0000-0000-0000-000000000004', '00d19df4-c72e-488d-8087-70b4fb18cc36', 'Crispy Pan-Fried Baltic Herrings', 'Golden butter-fried Baltic herrings with creamy mashed potatoes, clarified butter, and wild lingonberry compote.', 1650, 35, NOW(), NOW()),
    ('c1020000-0000-0000-0000-000000000005', '00d19df4-c72e-488d-8087-70b4fb18cc36', 'Lapland Cloudberry Panna Cotta', 'Silky vanilla pod panna cotta crowned with hand-picked golden Lapland cloudberries.', 850, 25, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock = EXCLUDED.stock,
    updated_at = NOW();

-- 7. Insert Products for Mario''s Artisan Pizza & Pasta (Store ID: 55ea6025-f6af-49b6-92af-7826844fad06)
INSERT INTO products (id, store_id, name, description, price, stock, created_at, updated_at)
VALUES
    ('c1030000-0000-0000-0000-000000000001', '55ea6025-f6af-49b6-92af-7826844fad06', 'Truffle Tagliatelle al Burro', 'Hand-rolled fresh egg tagliatelle tossed in fragrant black truffle butter, 24-month Parmigiano-Reggiano, and chives.', 1890, 40, NOW(), NOW()),
    ('c1030000-0000-0000-0000-000000000002', '55ea6025-f6af-49b6-92af-7826844fad06', 'Slow-Cooked Beef Ragù Pappardelle', '8-hour braised beef shank in Chianti red wine, sweet mirepoix, and San Marzano tomato reduction.', 1950, 35, NOW(), NOW()),
    ('c1030000-0000-0000-0000-000000000003', '55ea6025-f6af-49b6-92af-7826844fad06', 'Prosciutto di Parma & Rucola Pizza', 'Fior di latte mozzarella, 24-month aged Prosciutto di Parma, wild baby arugula, shaved pecorino, and aged balsamic glaze.', 1790, 50, NOW(), NOW()),
    ('c1030000-0000-0000-0000-000000000004', '55ea6025-f6af-49b6-92af-7826844fad06', 'Crispy Sicilian Cannoli (2 pcs)', 'Handmade crunchy pastry tubes filled with sweetened sheep ricotta, crushed green pistachios, and candied orange peel.', 650, 30, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock = EXCLUDED.stock,
    updated_at = NOW();

-- 8. Insert Products for Pamukkale Mediterranean Grill (Store ID: 9df1d050-2e8b-4113-87a2-07b52423dff5)
INSERT INTO products (id, store_id, name, description, price, stock, created_at, updated_at)
VALUES
    ('c1040000-0000-0000-0000-000000000001', '9df1d050-2e8b-4113-87a2-07b52423dff5', 'Charcoal Chicken Shish Platter', 'Marinated chicken breast skewers grilled over natural wood charcoal, served with fragrant bulgur pilaf, sumac onions, and tzatziki.', 1590, 45, NOW(), NOW()),
    ('c1040000-0000-0000-0000-000000000002', '9df1d050-2e8b-4113-87a2-07b52423dff5', 'Handmade Lamb Beyti Wrap', 'Spiced minced lamb kebab wrapped in thin lavash bread, baked and sliced, drizzled with hot tomato-butter sauce and strained yogurt.', 1750, 35, NOW(), NOW()),
    ('c1040000-0000-0000-0000-000000000003', '9df1d050-2e8b-4113-87a2-07b52423dff5', 'Crispy Falafel & Hummus Mezze', 'Golden herbaceous chickpea falafels on a bed of silky tahini hummus, fresh tabbouleh, pickled turnip, and warm pita.', 1390, 50, NOW(), NOW()),
    ('c1040000-0000-0000-0000-000000000004', '9df1d050-2e8b-4113-87a2-07b52423dff5', 'Antep Pistachio Baklava (4 pcs)', 'Forty layers of razor-thin handmade phyllo pastry soaked in lemon honey syrup, packed with double-roasted Gaziantep pistachios.', 700, 40, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock = EXCLUDED.stock,
    updated_at = NOW();

-- 9. Insert Products for Ravintola Kuurna & Nordic Bistro (Store ID: b1000000-0000-0000-0000-000000000001)
INSERT INTO products (id, store_id, name, description, price, stock, created_at, updated_at)
VALUES
    ('c1050000-0000-0000-0000-000000000001', 'b1000000-0000-0000-0000-000000000001', 'Slow-Roasted Lapland Reindeer Tenderloin', 'Finnish Lapland reindeer with velvety parsnip purée, sautéed yellow chanterelle mushrooms, and juniper berry red wine reduction.', 2850, 30, NOW(), NOW()),
    ('c1050000-0000-0000-0000-000000000002', 'b1000000-0000-0000-0000-000000000001', 'Charred Arctic Char Fillet', 'Pan-seared wild Arctic char from Lake Inari with smoked butter sauce, roasted baby beets, and crispy buttered dill potatoes.', 2400, 35, NOW(), NOW()),
    ('c1050000-0000-0000-0000-000000000003', 'b1000000-0000-0000-0000-000000000001', 'Yellow Chanterelle & Barley Risotto', 'Creamy pearl barley risotto infused with foraged Finnish porcini and yellow chanterelles, finished with aged Västerbotten cheese.', 1980, 40, NOW(), NOW()),
    ('c1050000-0000-0000-0000-000000000004', 'b1000000-0000-0000-0000-000000000001', 'Traditional Karelian Pies & Egg Butter (3 pcs)', 'Crispy rye crust pies filled with silky rice pudding, served warm with homemade churned salted egg butter.', 850, 50, NOW(), NOW()),
    ('c1050000-0000-0000-0000-000000000005', 'b1000000-0000-0000-0000-000000000001', 'Wild Sea Buckthorn & White Chocolate Tart', 'Tangy Finnish sea buckthorn curd in a crisp sweet pastry shell with toasted meringue and white chocolate snow.', 950, 25, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock = EXCLUDED.stock,
    updated_at = NOW();

-- 10. Insert Products for Friends & Brgrs Helsinki (Store ID: b2000000-0000-0000-0000-000000000002)
INSERT INTO products (id, store_id, name, description, price, stock, created_at, updated_at)
VALUES
    ('c1060000-0000-0000-0000-000000000001', 'b2000000-0000-0000-0000-000000000002', 'Classic Friends Smash Burger', 'Double smashed Finnish beef patty, melted mature cheddar, house mayonnaise, fresh butterhead lettuce, and pickled gherkins on a fresh brioche bun.', 1390, 60, NOW(), NOW()),
    ('c1060000-0000-0000-0000-000000000002', 'b2000000-0000-0000-0000-000000000002', 'Black Truffle & Bacon Burger', 'Double Finnish smash beef patties, smoked pork bacon, melted Swiss Emmental, slow-caramelized onions, and black truffle aioli.', 1590, 50, NOW(), NOW()),
    ('c1060000-0000-0000-0000-000000000003', 'b2000000-0000-0000-0000-000000000002', 'Crispy Buttermilk Chicken Burger', 'Golden fried seasoned buttermilk chicken fillet, fiery sriracha red cabbage slaw, pickled jalapeños, and fresh herb mayo.', 1450, 45, NOW(), NOW()),
    ('c1060000-0000-0000-0000-000000000004', 'b2000000-0000-0000-0000-000000000002', 'Triple-Cooked Hand-Cut Fries', 'Finnish potatoes cut fresh daily in-house, triple-cooked to crispy perfection, dusted with sea salt and served with garlic mayo dip.', 520, 100, NOW(), NOW()),
    ('c1060000-0000-0000-0000-000000000005', 'b2000000-0000-0000-0000-000000000002', 'Artisan Salted Caramel Milkshake', 'Creamy Madagascar vanilla bean ice cream blended with homemade sea-salted caramel sauce, crowned with whipped cream.', 650, 40, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock = EXCLUDED.stock,
    updated_at = NOW();

-- 11. Insert Products for Tokyo Street & Ramen House (Store ID: b3000000-0000-0000-0000-000000000003)
INSERT INTO products (id, store_id, name, description, price, stock, created_at, updated_at)
VALUES
    ('c1070000-0000-0000-0000-000000000001', 'b3000000-0000-0000-0000-000000000003', 'Hakata Tonkotsu Special Ramen', '16-hour simmered pork bone broth, handcrafted springy noodles, tender melt-in-mouth chashu pork, seasoned soft ajitsuke egg, wood ear mushrooms, and fresh scallions.', 1680, 50, NOW(), NOW()),
    ('c1070000-0000-0000-0000-000000000002', 'b3000000-0000-0000-0000-000000000003', 'Spicy Sapporo Miso Ramen', 'Roasted red miso & chicken broth, aromatic chili oil, spicy ground pork, sweet Hokkaido corn, crispy bamboo shoots, and soft-boiled egg.', 1750, 45, NOW(), NOW()),
    ('c1070000-0000-0000-0000-000000000003', 'b3000000-0000-0000-0000-000000000003', 'Pan-Fried Pork & Scallion Gyoza (6 pcs)', 'Hand-pinched Japanese dumplings pan-crisped with a golden lace skirt, served with seasoned soy-sesame vinegar dip.', 890, 60, NOW(), NOW()),
    ('c1070000-0000-0000-0000-000000000004', 'b3000000-0000-0000-0000-000000000003', 'Crispy Chicken Karaage Rice Bowl', 'Japanese double-fried ginger-soy chicken thigh bites over fluffy Koshihikari rice, Kewpie mayonnaise, togarashi seven-spice, and sesame slaw.', 1490, 40, NOW(), NOW()),
    ('c1070000-0000-0000-0000-000000000005', 'b3000000-0000-0000-0000-000000000003', 'Matcha Dorayaki Pancakes (2 pcs)', 'Fluffy Japanese honey pancakes sandwiched with authentic Uji matcha cream and sweet azuki red bean paste.', 620, 35, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock = EXCLUDED.stock,
    updated_at = NOW();

-- 12. Insert Products for Green Hippo Cafe & Bowls (Store ID: b4000000-0000-0000-0000-000000000004)
INSERT INTO products (id, store_id, name, description, price, stock, created_at, updated_at)
VALUES
    ('c1080000-0000-0000-0000-000000000001', 'b4000000-0000-0000-0000-000000000004', 'Signature Avocado & Poached Egg Toast', 'Chunky smashed Hass avocado, two organic soft poached eggs, Aleppo chili flakes, toasted pumpkin seeds, and creamy Danish feta on artisan sourdough.', 1290, 50, NOW(), NOW()),
    ('c1080000-0000-0000-0000-000000000002', 'b4000000-0000-0000-0000-000000000004', 'Warm Teriyaki Salmon Poke Bowl', 'Sesame-crusted seared salmon, steamed edamame, fresh avocado, pickled red cabbage, wakame seaweed salad, and warm brown sushi rice with teriyaki glaze.', 1790, 45, NOW(), NOW()),
    ('c1080000-0000-0000-0000-000000000003', 'b4000000-0000-0000-0000-000000000004', 'Organic Amazonian Acai Smoothie Bowl', 'Pure organic acai blended thick with frozen berries and banana, topped with roasted coconut-almond granola, chia seeds, and fresh strawberries.', 1050, 40, NOW(), NOW()),
    ('c1080000-0000-0000-0000-000000000004', 'b4000000-0000-0000-0000-000000000004', 'Grilled Halloumi Rainbow Mezze Salad', 'Charred golden Cypriot halloumi cheese, roasted sweet potato wedges, organic baby spinach, pomegranate pearls, and lemon-tahini dressing.', 1420, 40, NOW(), NOW()),
    ('c1080000-0000-0000-0000-000000000005', 'b4000000-0000-0000-0000-000000000004', 'Finnish Cinnamon Bun (Korvapuusti) & Oat Latte', 'Freshly baked traditional cardamom-cinnamon scroll paired with a double-shot specialty espresso and velvety Finnish oat milk.', 780, 50, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock = EXCLUDED.stock,
    updated_at = NOW();

COMMIT;
